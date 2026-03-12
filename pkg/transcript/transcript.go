package transcript

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
)

var (
	ReApiKey = regexp.MustCompile(`"INNERTUBE_API_KEY":"([^"]+)"`)
	ReApiKeyEscaped = regexp.MustCompile(`INNERTUBE_API_KEY\\":\\"([^\\"]+)\\"`)
)

type YoutubeTranscript struct {
	Config *TranscriptConfig
}

func NewYoutubeTranscript(config *TranscriptConfig) *YoutubeTranscript {
	if config == nil {
		config = &TranscriptConfig{}
	}
	return &YoutubeTranscript{Config: config}
}

func (yt *YoutubeTranscript) FetchTranscript(ctx context.Context, videoId string) ([]TranscriptResponse, error) {
	identifier, err := RetrieveVideoId(videoId)
	if err != nil {
		return nil, err
	}

	lang := yt.Config.Lang
	userAgent := yt.Config.UserAgent
	if userAgent == "" {
		userAgent = DefaultUserAgent
	}

	cacheKey := fmt.Sprintf("yt:transcript:%s:%s", identifier, lang)
	if yt.Config.Cache != nil {
		cached, err := yt.Config.Cache.Get(ctx, cacheKey)
		if err == nil && cached != "" {
			var transcript []TranscriptResponse
			if err := json.Unmarshal([]byte(cached), &transcript); err == nil {
				return transcript, nil
			}
		}
	}

	protocol := "https"
	if yt.Config.DisableHTTPS {
		protocol = "http"
	}

	watchUrl := fmt.Sprintf("%s://www.youtube.com/watch?v=%s", protocol, identifier)
	
	var videoRes *http.Response
	if yt.Config.VideoFetch != nil {
		videoRes, err = yt.Config.VideoFetch(ctx, watchUrl, lang, userAgent)
	} else {
		videoRes, err = DefaultFetch(ctx, watchUrl, "GET", lang, userAgent, nil, nil)
	}

	if err != nil {
		return nil, err
	}
	defer videoRes.Body.Close()

	if videoRes.StatusCode != http.StatusOK {
		return nil, &VideoUnavailableError{VideoID: identifier}
	}

	bodyBytes, err := io.ReadAll(videoRes.Body)
	if err != nil {
		return nil, err
	}
	body := string(bodyBytes)

	if regexp.MustCompile(`class="g-recaptcha"`).MatchString(body) {
		return nil, &TooManyRequestError{}
	}

	apiKeyMatch := ReApiKey.FindStringSubmatch(body)
	if apiKeyMatch == nil {
		apiKeyMatch = ReApiKeyEscaped.FindStringSubmatch(body)
	}

	if apiKeyMatch == nil {
		return nil, &NotAvailableError{VideoID: identifier}
	}
	apiKey := apiKeyMatch[1]

	playerEndpoint := fmt.Sprintf("%s://www.youtube.com/youtubei/v1/player?key=%s", protocol, apiKey)
	playerBody := map[string]interface{}{
		"context": map[string]interface{}{
			"client": map[string]interface{}{
				"clientName":    "ANDROID",
				"clientVersion": "20.10.38",
			},
		},
		"videoId": identifier,
	}
	playerBodyBytes, _ := json.Marshal(playerBody)

	var playerRes *http.Response
	if yt.Config.PlayerFetch != nil {
		playerRes, err = yt.Config.PlayerFetch(ctx, playerEndpoint, "POST", lang, userAgent, playerBodyBytes, map[string]string{"Content-Type": "application/json"})
	} else {
		playerRes, err = DefaultFetch(ctx, playerEndpoint, "POST", lang, userAgent, playerBodyBytes, map[string]string{"Content-Type": "application/json"})
	}

	if err != nil {
		return nil, err
	}
	defer playerRes.Body.Close()

	if playerRes.StatusCode != http.StatusOK {
		return nil, &VideoUnavailableError{VideoID: identifier}
	}

	var playerJson map[string]interface{}
	if err := json.NewDecoder(playerRes.Body).Decode(&playerJson); err != nil {
		return nil, err
	}

	captions, ok := playerJson["captions"].(map[string]interface{})
	var tracklist map[string]interface{}
	if ok {
		tracklist, _ = captions["playerCaptionsTracklistRenderer"].(map[string]interface{})
	} else {
		tracklist, _ = playerJson["playerCaptionsTracklistRenderer"].(map[string]interface{})
	}

	playabilityStatus, _ := playerJson["playabilityStatus"].(map[string]interface{})
	isPlayableOk := playabilityStatus != nil && playabilityStatus["status"] == "OK"

	if captions == nil || tracklist == nil {
		if isPlayableOk {
			return nil, &DisabledError{VideoID: identifier}
		}
		return nil, &NotAvailableError{VideoID: identifier}
	}

	tracksRaw, _ := tracklist["captionTracks"].([]interface{})
	if len(tracksRaw) == 0 {
		return nil, &DisabledError{VideoID: identifier}
	}

	var selectedTrack map[string]interface{}
	if lang != "" {
		for _, t := range tracksRaw {
			track := t.(map[string]interface{})
			if track["languageCode"] == lang {
				selectedTrack = track
				break
			}
		}
	} else {
		selectedTrack = tracksRaw[0].(map[string]interface{})
	}

	if selectedTrack == nil {
		var available []string
		for _, t := range tracksRaw {
			track := t.(map[string]interface{})
			if l, ok := track["languageCode"].(string); ok {
				available = append(available, l)
			}
		}
		return nil, &NotAvailableLanguageError{VideoID: identifier, RequestedLang: lang, AvailableLangs: available}
	}

	transcriptURL, _ := selectedTrack["baseUrl"].(string)
	if transcriptURL == "" {
		transcriptURL, _ = selectedTrack["url"].(string)
	}
	if transcriptURL == "" {
		return nil, &NotAvailableError{VideoID: identifier}
	}

	transcriptURL = regexp.MustCompile(`&fmt=[^&]+`).ReplaceAllString(transcriptURL, "")
	if yt.Config.DisableHTTPS {
		transcriptURL = regexp.MustCompile(`^https://`).ReplaceAllString(transcriptURL, "http://")
	}

	var transRes *http.Response
	if yt.Config.TranscriptFetch != nil {
		transRes, err = yt.Config.TranscriptFetch(ctx, transcriptURL, lang, userAgent)
	} else {
		transRes, err = DefaultFetch(ctx, transcriptURL, "GET", lang, userAgent, nil, nil)
	}

	if err != nil {
		return nil, err
	}
	defer transRes.Body.Close()

	if transRes.StatusCode != http.StatusOK {
		if transRes.StatusCode == http.StatusTooManyRequests {
			return nil, &TooManyRequestError{}
		}
		return nil, &NotAvailableError{VideoID: identifier}
	}

	transBodyBytes, err := io.ReadAll(transRes.Body)
	if err != nil {
		return nil, err
	}
	transBody := string(transBodyBytes)

	results := ReXmlTranscript.FindAllStringSubmatch(transBody, -1)
	var transcript []TranscriptResponse
	for _, m := range results {
		if len(m) < 4 {
			continue
		}
		duration, _ := strconv.ParseFloat(m[2], 64)
		offset, _ := strconv.ParseFloat(m[1], 64)
		transcript = append(transcript, TranscriptResponse{
			Text:     DecodeXmlEntities(m[3]),
			Duration: duration,
			Offset:   offset,
			Lang:     func() string { if lang != "" { return lang }; return selectedTrack["languageCode"].(string) }(),
		})
	}

	if len(transcript) == 0 {
		return nil, &NotAvailableError{VideoID: identifier}
	}

	if yt.Config.Cache != nil {
		transcriptJSON, _ := json.Marshal(transcript)
		_ = yt.Config.Cache.Set(ctx, cacheKey, string(transcriptJSON), yt.Config.CacheTTL)
	}

	return transcript, nil
}

func FetchTranscript(ctx context.Context, videoId string, config *TranscriptConfig) ([]TranscriptResponse, error) {
	return NewYoutubeTranscript(config).FetchTranscript(ctx, videoId)
}

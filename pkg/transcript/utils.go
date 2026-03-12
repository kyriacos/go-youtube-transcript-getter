package transcript

import (
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"
)

var (
	DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
	ReYoutube        = regexp.MustCompile(`(?i)(?:v=|\/|v\/|embed\/|watch\?.*v=|youtu\.be\/|\/v\/|e\/|watch\?.*vi?=|\/embed\/|\/v\/|vi?\/|watch\?.*vi?=|youtu\.be\/|\/vi?\/|\/e\/)([a-zA-Z0-9_-]{11})`)
	ReVideoID        = regexp.MustCompile(`^[a-zA-Z0-9_-]{11}$`)
	ReXmlTranscript  = regexp.MustCompile(`<text start="([^"]*)" dur="([^"]*)">([^<]*)<\/text>`)
)

func RetrieveVideoId(videoId string) (string, error) {
	if ReVideoID.MatchString(videoId) {
		return videoId, nil
	}
	matches := ReYoutube.FindStringSubmatch(videoId)
	if len(matches) > 1 {
		return matches[1], nil
	}
	return "", &InvalidVideoIdError{}
}

var xmlEntities = map[string]string{
	"&amp;":  "&",
	"&lt;":   "<",
	"&gt;":   ">",
	"&quot;": "\"",
	"&#39;":  "'",
	"&apos;": "'",
}

func DecodeXmlEntities(text string) string {
	for entity, replacement := range xmlEntities {
		text = strings.ReplaceAll(text, entity, replacement)
	}
	return text
}

func DefaultFetch(ctx context.Context, url string, method string, lang string, userAgent string, body []byte, headers map[string]string) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = strings.NewReader(string(body))
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	if userAgent == "" {
		userAgent = DefaultUserAgent
	}
	req.Header.Set("User-Agent", userAgent)
	if lang != "" {
		req.Header.Set("Accept-Language", lang)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	return client.Do(req)
}

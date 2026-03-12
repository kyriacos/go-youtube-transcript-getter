package transcript

import "fmt"

type YoutubeTranscriptError struct {
	Message string
}

func (e *YoutubeTranscriptError) Error() string {
	return e.Message
}

type VideoUnavailableError struct {
	VideoID string
}

func (e *VideoUnavailableError) Error() string {
	return fmt.Sprintf("YoutubeTranscript - Video (%s) is unavailable", e.VideoID)
}

type TooManyRequestError struct{}

func (e *TooManyRequestError) Error() string {
	return "YoutubeTranscript - Too many requests (reCaptcha detected)"
}

type DisabledError struct {
	VideoID string
}

func (e *DisabledError) Error() string {
	return fmt.Sprintf("YoutubeTranscript - Transcripts are disabled for this video (%s)", e.VideoID)
}

type NotAvailableError struct {
	VideoID string
}

func (e *NotAvailableError) Error() string {
	return fmt.Sprintf("YoutubeTranscript - No transcripts available for this video (%s)", e.VideoID)
}

type NotAvailableLanguageError struct {
	VideoID         string
	RequestedLang   string
	AvailableLangs []string
}

func (e *NotAvailableLanguageError) Error() string {
	return fmt.Sprintf("YoutubeTranscript - No transcripts available in %s for this video (%s). Available: %v", e.RequestedLang, e.VideoID, e.AvailableLangs)
}

type InvalidVideoIdError struct{}

func (e *InvalidVideoIdError) Error() string {
	return "YoutubeTranscript - Invalid video ID"
}

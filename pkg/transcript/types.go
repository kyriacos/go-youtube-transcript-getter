package transcript

import (
	"context"
	"net/http"
	"time"
)

type TranscriptResponse struct {
	Text     string  `json:"text"`
	Duration float64 `json:"duration"`
	Offset   float64 `json:"offset"`
	Lang     string  `json:"lang"`
}

type TranscriptConfig struct {
	Lang            string
	UserAgent       string
	DisableHTTPS    bool
	VideoFetch      func(ctx context.Context, url string, lang string, userAgent string) (*http.Response, error)
	PlayerFetch     func(ctx context.Context, url string, method string, lang string, userAgent string, body []byte, headers map[string]string) (*http.Response, error)
	TranscriptFetch func(ctx context.Context, url string, lang string, userAgent string) (*http.Response, error)
	Cache           CacheStrategy
	CacheTTL        time.Duration
}

type CacheStrategy interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
}

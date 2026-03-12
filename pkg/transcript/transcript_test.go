package transcript

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func loadFixture(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("fixtures", name)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to load fixture %s: %v", name, err)
	}
	return string(content)
}

func loadJsonFixture(t *testing.T, name string, target interface{}) {
	t.Helper()
	content := loadFixture(t, name)
	if err := json.Unmarshal([]byte(content), target); err != nil {
		t.Fatalf("Failed to unmarshal fixture %s: %v", name, err)
	}
}

func TestYoutubeTranscript_FetchTranscript_Success(t *testing.T) {
	mux := http.NewServeMux()
	videoId := "TESTVIDEOID"
	apiKey := "test-key"

	mux.HandleFunc("/watch", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("v") != videoId {
			t.Errorf("Expected video ID %s, got %s", videoId, r.URL.Query().Get("v"))
		}
		fmt.Fprint(w, loadFixture(t, "watch.html"))
	})

	mux.HandleFunc("/youtubei/v1/player", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("key") != apiKey {
			t.Errorf("Expected API key %s, got %s", apiKey, r.URL.Query().Get("key"))
		}
		fmt.Fprint(w, loadFixture(t, "player-success.json"))
	})

	mux.HandleFunc("/api/timedtext", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, loadFixture(t, "transcript.xml"))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	// For simplicity in testing, we use a custom fetcher that redirects to our test server
	config := &TranscriptConfig{
		VideoFetch: func(ctx context.Context, url string, lang string, userAgent string) (*http.Response, error) {
			return http.DefaultClient.Get(server.URL + "/watch?v=" + videoId)
		},
		PlayerFetch: func(ctx context.Context, url string, method string, lang string, userAgent string, body []byte, headers map[string]string) (*http.Response, error) {
			return http.DefaultClient.Post(server.URL+"/youtubei/v1/player?key="+apiKey, "application/json", nil)
		},
		TranscriptFetch: func(ctx context.Context, url string, lang string, userAgent string) (*http.Response, error) {
			return http.DefaultClient.Get(server.URL + "/api/timedtext")
		},
	}

	yt := NewYoutubeTranscript(config)
	transcript, err := yt.FetchTranscript(context.Background(), videoId)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []TranscriptResponse{
		{Text: "Hello world", Duration: 1.5, Offset: 0, Lang: "en"},
		{Text: "Second line", Duration: 2.0, Offset: 1.5, Lang: "en"},
	}

	if !reflect.DeepEqual(transcript, expected) {
		t.Errorf("Expected %+v, got %+v", expected, transcript)
	}
}

func TestYoutubeTranscript_FetchTranscript_Errors(t *testing.T) {
	videoId := "TESTVIDEOID"

	t.Run("Video Unavailable", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		config := &TranscriptConfig{
			VideoFetch: func(ctx context.Context, url string, lang string, userAgent string) (*http.Response, error) {
				return http.DefaultClient.Get(server.URL)
			},
		}

		yt := NewYoutubeTranscript(config)
		_, err := yt.FetchTranscript(context.Background(), videoId)
		if _, ok := err.(*VideoUnavailableError); !ok {
			t.Errorf("Expected VideoUnavailableError, got %T: %v", err, err)
		}
	})

	t.Run("Too Many Requests", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, loadFixture(t, "watch-recaptcha.html"))
		}))
		defer server.Close()

		config := &TranscriptConfig{
			VideoFetch: func(ctx context.Context, url string, lang string, userAgent string) (*http.Response, error) {
				return http.DefaultClient.Get(server.URL)
			},
		}

		yt := NewYoutubeTranscript(config)
		_, err := yt.FetchTranscript(context.Background(), videoId)
		if _, ok := err.(*TooManyRequestError); !ok {
			t.Errorf("Expected TooManyRequestError, got %T: %v", err, err)
		}
	})

	t.Run("Transcripts Disabled", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/watch" {
				fmt.Fprint(w, loadFixture(t, "watch.html"))
			} else {
				fmt.Fprint(w, loadFixture(t, "player-disabled.json"))
			}
		}))
		defer server.Close()

		config := &TranscriptConfig{
			VideoFetch: func(ctx context.Context, url string, lang string, userAgent string) (*http.Response, error) {
				return http.DefaultClient.Get(server.URL + "/watch")
			},
			PlayerFetch: func(ctx context.Context, url string, method string, lang string, userAgent string, body []byte, headers map[string]string) (*http.Response, error) {
				return http.DefaultClient.Get(server.URL + "/player")
			},
		}

		yt := NewYoutubeTranscript(config)
		_, err := yt.FetchTranscript(context.Background(), videoId)
		if _, ok := err.(*DisabledError); !ok {
			t.Errorf("Expected DisabledError, got %T: %v", err, err)
		}
	})
}

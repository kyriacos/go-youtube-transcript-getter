package transcript

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRetrieveVideoId(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"3oXfZ-Rffz8", "3oXfZ-Rffz8", false},
		{"https://www.youtube.com/watch?v=3oXfZ-Rffz8", "3oXfZ-Rffz8", false},
		{"https://youtu.be/3oXfZ-Rffz8", "3oXfZ-Rffz8", false},
		{"invalid", "", true},
		{"https://example.com", "", true},
		{"../.././../.", "", true},
		{"hello world", "", true},
		{"abc!@#$%^&*", "", true},
		{"abc_def-123", "abc_def-123", false},
		{"___________", "___________", false},
		{"-----------", "-----------", false},
	}

	for _, tt := range tests {
		got, err := RetrieveVideoId(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("RetrieveVideoId(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if got != tt.expected {
			t.Errorf("RetrieveVideoId(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestDecodeXmlEntities(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"rock &amp; roll", "rock & roll"},
		{"it&#39;s", "it's"},
		{"it&apos;s", "it's"},
		{"a &quot;test&quot;", "a \"test\""},
		{"&lt;tag&gt;", "<tag>"},
		{"A &amp; B &lt; C &gt; D", "A & B < C > D"},
		{"Hello world", "Hello world"},
	}

	for _, tt := range tests {
		got := DecodeXmlEntities(tt.input)
		if got != tt.expected {
			t.Errorf("DecodeXmlEntities(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestDefaultFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("User-Agent header is missing")
		}
		if r.Method == "POST" {
			body, _ := io.ReadAll(r.Body)
			if string(body) == "" {
				t.Error("POST body is missing")
			}
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx := context.Background()

	t.Run("GET request", func(t *testing.T) {
		res, err := DefaultFetch(ctx, server.URL, "GET", "en", "TestAgent", nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected status OK, got %v", res.Status)
		}
	})

	t.Run("POST request with body", func(t *testing.T) {
		body := []byte(`{"test":"data"}`)
		res, err := DefaultFetch(ctx, server.URL, "POST", "en", "TestAgent", body, map[string]string{"Content-Type": "application/json"})
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != http.StatusOK {
			t.Errorf("Expected status OK, got %v", res.Status)
		}
	})
}

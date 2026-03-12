# go-youtube-transcript-getter

A Go library and CLI tool to fetch transcripts from YouTube videos. 

It uses YouTube's unofficial API (Innertube) to discover and fetch caption tracks.

#### This is a port of the [youtube-transcript-plus](https://github.com/ericmmartin/youtube-transcript-plus) library.

## Installation

### As a Library

```bash
go get github.com/kyriacos/go-youtube-transcript-getter
```

### As a CLI Tool

```bash
go install github.com/kyriacos/go-youtube-transcript-getter/cmd/youtube-transcript-getter@latest
```

## Usage

### Basic Usage (Library)

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/kyriacos/go-youtube-transcript-getter/pkg/transcript"
)

func main() {
	videoID := "3oXfZ-Rffz8"
	
	// Fetch transcript using default settings
	results, err := transcript.FetchTranscript(context.Background(), videoID, nil)
	if err != nil {
		log.Fatal(err)
	}

	for _, line := range results {
		fmt.Printf("[%0.2f] %s\n", line.Offset, line.Text)
	}
}
```

### CLI Usage

```bash
# Basic usage
youtube-transcript-getter 3oXfZ-Rffz8

# Specify language
youtube-transcript-getter 3oXfZ-Rffz8 fr

# Output is JSON formatted
youtube-transcript-getter 3oXfZ-Rffz8 | jq .
```

### Configuration Options

You can customize the behavior by passing a `TranscriptConfig` struct:

```go
config := &transcript.TranscriptConfig{
    Lang:         "en",
    UserAgent:    "Custom User Agent",
    DisableHTTPS: false,
    Cache:        transcript.NewInMemoryCache(),
    CacheTTL:     1 * time.Hour,
}

yt := transcript.NewYoutubeTranscript(config)
results, err := yt.FetchTranscript(ctx, videoID)
```

### Caching

The library provides two built-in caching strategies:

#### In-Memory Cache

```go
cache := transcript.NewInMemoryCache()
config := &transcript.TranscriptConfig{
    Cache:    cache,
    CacheTTL: 30 * time.Minute,
}
```

#### File System Cache

```go
cache, _ := transcript.NewFsCache("./cache-dir")
config := &transcript.TranscriptConfig{
    Cache:    cache,
    CacheTTL: 24 * time.Hour,
}
```

### Custom Fetchers

You can intercept and modify HTTP requests by providing custom fetch functions in the config. This is useful for proxies, custom headers, or logging.

- `VideoFetch`: For fetching the initial watch page.
- `PlayerFetch`: For the Innertube player API (POST).
- `TranscriptFetch`: For the actual XML transcript data.

### Error Handling

The library returns specific error types for different failure scenarios:

- `VideoUnavailableError`: Video is removed or private.
- `DisabledError`: Transcripts are disabled for this video.
- `NotAvailableError`: No transcripts available.
- `NotAvailableLanguageError`: Requested language not found.
- `TooManyRequestError`: Rate limited (reCaptcha).
- `InvalidVideoIdError`: Provided ID or URL is invalid.

Example:
```go
results, err := transcript.FetchTranscript(ctx, videoID, nil)
if err != nil {
    if _, ok := err.(*transcript.TooManyRequestError); ok {
        fmt.Println("Rate limited by YouTube")
    }
}
```

## Features

- Support for video IDs and various YouTube URL formats.
- Built-in In-Memory and File System caching.
- Pluggable HTTP fetchers.
- Clean, idiomatic Go API.

## Special thanks
- This is a port of the [youtube-transcript-plus](https://github.com/ericmmartin/youtube-transcript-plus) library.
- Thank you!

## License

MIT

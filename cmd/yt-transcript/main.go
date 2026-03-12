package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/kyriacos/go-youtube-transcript-getter/pkg/transcript"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: yt-transcript <video-id> [lang]")
		os.Exit(1)
	}

	videoId := os.Args[1]
	lang := ""
	if len(os.Args) > 2 {
		lang = os.Args[2]
	}

	config := &transcript.TranscriptConfig{
		Lang: lang,
	}

	ctx := context.Background()
	results, err := transcript.FetchTranscript(ctx, videoId, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	output, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))
}

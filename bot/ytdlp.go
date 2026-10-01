package bot

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type SearchResult struct {
	Title    string  `json:"title"`
	URL      string  `json:"webpage_url"`
	Uploader string  `json:"uploader"`
	Duration float64 `json:"duration"`
}

type SongMetadata struct {
	Title      string  `json:"title"`
	Uploader   string  `json:"uploader"`
	Duration   float64 `json:"duration"`
	Thumbnail  string  `json:"thumbnail"`
	WebpageURL string  `json:"webpage_url"`
}

func FetchMetadata(url string) (*SongMetadata, error) {
	query := url
	if !strings.HasPrefix(query, "http://") && !strings.HasPrefix(query, "https://") {
		query = "ytsearch1:" + query
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "yt-dlp",
		"-j",
		"--no-playlist",
		"--no-warnings",
		"--quiet",
		"--extractor-args", "youtube:player_client=android,web",
		query,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("yt-dlp Timeout (>20s)")
		}
		return nil, fmt.Errorf("yt-dlp Error: %v | Details: %s", err, stderr.String())
	}
	var meta SongMetadata
	if err := json.Unmarshal(out, &meta); err != nil {
		return nil, fmt.Errorf("Error parsing JSON: %v", err)
	}
	return &meta, nil
}

func SearchTracks(query string) ([]SearchResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "yt-dlp",
		"ytsearch5:"+query,
		"-j",
		"--flat-playlist",
		"--no-warnings",
		"--ignore-errors",
		"--quiet",
		"--extractor-args", "youtube:player_client=android,web",
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("yt-dlp Search timeout")
		}
		return nil, fmt.Errorf("yt-dlp search error: %v | Details: %s", err, stderr.String())
	}
	var results []SearchResult
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var raw struct {
			Title    string  `json:"title"`
			URL      string  `json:"url"`
			WebURL   string  `json:"webpage_url"`
			ID       string  `json:"id"`
			Uploader string  `json:"uploader"`
			Duration float64 `json:"duration"`
		}
		if err := json.Unmarshal(line, &raw); err == nil {
			finalURL := raw.WebURL
			if finalURL == "" && raw.URL != "" {
				finalURL = raw.URL
			}
			if finalURL == "" && raw.ID != "" {
				finalURL = "https://www.youtube.com/watch?v=" + raw.ID
			}

			if finalURL != "" {
				results = append(results, SearchResult{
					Title:    raw.Title,
					URL:      finalURL,
					Uploader: raw.Uploader,
					Duration: raw.Duration,
				})
			}
		}
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("No results found.")
	}
	return results, nil
}

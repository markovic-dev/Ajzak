package bot

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
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
	cmd := exec.Command("yt-dlp", "-j", "--no-playlist", query)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp greška: %v", err)
	}
	var meta SongMetadata
	if err := json.Unmarshal(out, &meta); err != nil {
		return nil, fmt.Errorf("greška pri parsiranju JSON-a: %v", err)
	}
	return &meta, nil
}

func SearchTracks(query string) ([]SearchResult, error) {
	cmd := exec.Command("yt-dlp",
		"ytsearch5:"+query,
		"-j",
		"--flat-playlist",
		"--no-warnings",
		"--ignore-errors",
	)

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp search error: %v", err)
	}

	var results []SearchResult
	scanner := bufio.NewScanner(bytes.NewReader(out))

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		// Privremena struktura za parsiranje raznolikih yt-dlp JSON odgovora
		var raw struct {
			Title    string  `json:"title"`
			URL      string  `json:"url"`
			WebURL   string  `json:"webpage_url"`
			ID       string  `json:"id"`
			Uploader string  `json:"uploader"`
			Duration float64 `json:"duration"`
		}

		if err := json.Unmarshal(line, &raw); err == nil {
			// Osiguravamo ispravan YouTube URL
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
		return nil, fmt.Errorf("nema pronađenih rezultata")
	}

	return results, nil
}

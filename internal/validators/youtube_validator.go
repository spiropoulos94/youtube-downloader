package validators

import (
	"fmt"
	"net/url"
	"strings"
)

// YouTubeURLValidator implements URLValidatorInterface
type YouTubeURLValidator struct{}

// NewYouTubeURLValidator creates a new instance of YouTubeURLValidator
func NewYouTubeURLValidator() URLValidatorInterface {
	return &YouTubeURLValidator{}
}

// Validate checks if the given URL is a valid YouTube URL
func (v *YouTubeURLValidator) Validate(urlStr string) error {
	// Check if URL is empty
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	// Parse the URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Check if it's a YouTube URL
	if !strings.Contains(parsedURL.Host, "youtube.com") {
		return fmt.Errorf("not a YouTube URL")
	}

	// Check if it's a watch URL
	if parsedURL.Path != "/watch" {
		return fmt.Errorf("not a YouTube watch URL")
	}

	// Check for video ID
	videoID := parsedURL.Query().Get("v")
	if videoID == "" {
		return fmt.Errorf("missing video ID")
	}

	// Basic video ID validation (YouTube video IDs are typically 11 characters)
	if len(videoID) != 11 {
		return fmt.Errorf("invalid video ID length")
	}

	return nil
}

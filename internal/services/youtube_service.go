package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"spiropoulos94/youtube-downloader/internal/rediskeys"

	"github.com/redis/go-redis/v9"
)

type YouTubeService struct {
	outputDir string
	redis     *redis.Client
}

func NewYouTubeService(outputDir string, redis *redis.Client) *YouTubeService {
	return &YouTubeService{
		outputDir: outputDir,
		redis:     redis,
	}
}

// UpdateLastRequestTime updates the last request time for a file
func (s *YouTubeService) UpdateLastRequestTime(filePath string) error {
	ctx := context.Background()
	key := rediskeys.GetLastRequestKey(filePath)
	now := time.Now().Format(time.RFC3339)

	if err := s.redis.Set(ctx, key, now, 1*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to update last request time: %v", err)
	}
	return nil
}

// GetURLHash returns a hash of the URL that can be used to identify the video
func (s *YouTubeService) GetURLHash(url string) string {
	urlHash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(urlHash[:8]) // Use first 8 chars of hash
}

func (s *YouTubeService) DownloadVideo(url string) (string, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(s.outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %v", err)
	}

	// Check if yt-dlp is installed
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return "", fmt.Errorf("yt-dlp is not installed. Please install it first:\nOn macOS: brew install yt-dlp\nOn Linux: sudo apt install yt-dlp or sudo pip install yt-dlp")
	}

	// Get URL hash
	urlHashStr := s.GetURLHash(url)

	// Check if video already exists
	existingFiles, err := os.ReadDir(s.outputDir)
	if err != nil {
		return "", fmt.Errorf("failed to read output directory: %v", err)
	}

	// Look for existing video with the same URL hash
	for _, file := range existingFiles {
		if strings.HasSuffix(file.Name(), urlHashStr+".mp4") {
			filePath := filepath.Join(s.outputDir, file.Name())
			// Verify file exists and is readable
			if _, err := os.Stat(filePath); err == nil {
				// update last request time
				if err := s.UpdateLastRequestTime(filePath); err != nil {
					return "", err
				}
				// Return the existing file path
				return filePath, nil
			}
		}
	}

	// If video doesn't exist, proceed with download
	outputTemplate := filepath.Join(s.outputDir, fmt.Sprintf("%%(title)s_%s.%%(ext)s", urlHashStr))
	cmd := exec.Command("yt-dlp",
		"-o", outputTemplate,
		"--merge-output-format", "mp4", // Force output to be MP4
		"--windows-filenames", // Only restrict characters that are illegal in Windows
		"--no-playlist",
		"--progress",  // Show download progress
		"--newline",   // Force progress on new lines
		"--no-colors", // Disable colors in output
		url)

	// Set up command output
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to download video: %v", err)
	}

	// Get the downloaded file path
	files, err := os.ReadDir(s.outputDir)
	if err != nil {
		return "", fmt.Errorf("failed to read output directory: %v", err)
	}

	// Find the file with our URL hash
	for _, file := range files {
		if strings.HasSuffix(file.Name(), urlHashStr+".mp4") {
			filePath := filepath.Join(s.outputDir, file.Name())
			// Update last request time for the newly downloaded file
			if err := s.UpdateLastRequestTime(filePath); err != nil {
				return "", err
			}
			return filePath, nil
		}
	}

	return "", fmt.Errorf("no downloaded file found with hash %s", urlHashStr)
}

// GetOriginalFilename extracts the original title from a downloaded video filename.
// The format is expected to be "title_hash.mp4", and this function returns "title.mp4".
// If escape is true, it also escapes special characters for use in Content-Disposition headers.
func (s *YouTubeService) GetOriginalFilename(filePath string, escape bool) string {
	fileName := filepath.Base(filePath)
	// Remove the hash part to get just the title portion
	lastUnderscore := strings.LastIndex(fileName, "_")
	titlePart := fileName
	if lastUnderscore != -1 {
		titlePart = fileName[:lastUnderscore]
	}
	// Ensure the extension is preserved
	if !strings.HasSuffix(titlePart, ".mp4") {
		titlePart += ".mp4"
	}

	if escape {
		// Escape double quotes for Content-Disposition header
		titlePart = strings.ReplaceAll(titlePart, `"`, `\"`)
	}

	return titlePart
}

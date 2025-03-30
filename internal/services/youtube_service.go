package services

import (
	"context"
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

func (s *YouTubeService) DownloadVideo(url string) (string, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(s.outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %v", err)
	}

	// Check if yt-dlp is installed
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return "", fmt.Errorf("yt-dlp is not installed. Please install it first:\nOn macOS: brew install yt-dlp\nOn Linux: sudo apt install yt-dlp or sudo pip install yt-dlp")
	}

	// First, try to get the video title without downloading
	titleCmd := exec.Command("yt-dlp",
		"--get-title",
		"--no-playlist",
		url)

	titleOutput, err := titleCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get video title: %v", err)
	}

	// Clean up the title (remove newlines and trim spaces)
	title := strings.TrimSpace(string(titleOutput))

	// Check if video already exists
	existingFiles, err := os.ReadDir(s.outputDir)
	if err != nil {
		return "", fmt.Errorf("failed to read output directory: %v", err)
	}

	// Look for existing video with the same title
	for _, file := range existingFiles {
		if strings.Contains(file.Name(), title) && (strings.HasSuffix(file.Name(), ".mp4") || strings.HasSuffix(file.Name(), ".webm")) {
			filePath := filepath.Join(s.outputDir, file.Name())
			// update last request time
			if err := s.UpdateLastRequestTime(filePath); err != nil {
				return "", err
			}
			//  and return the file path
			return filePath, nil
		}
	}

	// If video doesn't exist, proceed with download
	cmd := exec.Command("yt-dlp",
		"-o", filepath.Join(s.outputDir, "%(title)s.%(ext)s"),
		"--merge-output-format", "mp4", // Force output to be MP4
		"--no-playlist",
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

	// Find the most recently created file
	var latestFile string
	var latestTime int64
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Unix() > latestTime {
			latestTime = info.ModTime().Unix()
			latestFile = filepath.Join(s.outputDir, info.Name())
		}
	}

	if latestFile == "" {
		return "", fmt.Errorf("no downloaded file found")
	}

	// Update last request time for the newly downloaded file
	if err := s.UpdateLastRequestTime(latestFile); err != nil {
		return "", err
	}
	return latestFile, nil
}

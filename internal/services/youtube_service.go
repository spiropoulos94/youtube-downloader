package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type YouTubeService struct {
	outputDir string
	redis     *redis.Client
	mu        sync.RWMutex
}

func NewYouTubeService(outputDir string, redis *redis.Client) *YouTubeService {
	return &YouTubeService{
		outputDir: outputDir,
		redis:     redis,
	}
}

// getRefCountKey returns the Redis key for a file's reference count
func (s *YouTubeService) getRefCountKey(filePath string) string {
	return fmt.Sprintf("video:refcount:%s", filePath)
}

// IncrementRefCount increases the reference count for a video file
func (s *YouTubeService) IncrementRefCount(filePath string) error {
	ctx := context.Background()
	key := s.getRefCountKey(filePath)

	log.Printf("Incrementing reference count for file: %s", filePath)

	// Use Redis INCR to atomically increment the count
	count, err := s.redis.Incr(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to increment reference count: %v", err)
	}

	// Set expiration to 24 hours to prevent stale entries
	if count == 1 {
		s.redis.Expire(ctx, key, 24*time.Hour)
	}

	return nil
}

// DecrementRefCount decreases the reference count for a video file and deletes it if count reaches zero
func (s *YouTubeService) DecrementRefCount(filePath string) error {
	ctx := context.Background()
	key := s.getRefCountKey(filePath)

	log.Printf("Decrementing reference count for file: %s", filePath)

	// Use Redis DECR to atomically decrement the count
	count, err := s.redis.Decr(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to decrement reference count: %v", err)
	}

	log.Printf("Reference count for file: %s is now %d", filePath, count)

	if count <= 0 {
		// Delete the Redis key
		if err := s.redis.Del(ctx, key).Err(); err != nil {
			return fmt.Errorf("failed to delete Redis key: %v", err)
		}

		// Delete the file
		if err := os.Remove(filePath); err != nil {
			return fmt.Errorf("failed to delete file: %v", err)
		}
	}

	return nil
}

// GetRefCount returns the current reference count for a file
// This method can be used for monitoring purposes in the future
// e.g. to see how many users are downloading the same file
func (s *YouTubeService) GetRefCount(filePath string) (int64, error) {
	ctx := context.Background()
	key := s.getRefCountKey(filePath)

	count, err := s.redis.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get reference count: %v", err)
	}

	return count, nil
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
			if err := s.IncrementRefCount(filePath); err != nil {
				return "", err
			}
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

	// Increment reference count for the newly downloaded file
	if err := s.IncrementRefCount(latestFile); err != nil {
		return "", err
	}
	return latestFile, nil
}

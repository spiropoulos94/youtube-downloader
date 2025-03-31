package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"spiropoulos94/youtube-downloader/internal/rediskeys"

	"github.com/redis/go-redis/v9"
)

type CleanupService struct {
	outputDir string
	redis     *redis.Client
	stopChan  chan struct{}
}

func NewCleanupService(outputDir string, redis *redis.Client) *CleanupService {
	return &CleanupService{
		outputDir: outputDir,
		redis:     redis,
		stopChan:  make(chan struct{}),
	}
}

// Start begins the cleanup service
func (s *CleanupService) Start() {
	go s.runCleanupLoop()
}

// Stop gracefully stops the cleanup service
func (s *CleanupService) Stop() {
	close(s.stopChan)
}

// runCleanupLoop runs the cleanup process every 5 seconds
func (s *CleanupService) runCleanupLoop() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			if err := s.cleanup(); err != nil {
				log.Printf("Error during cleanup: %v", err)
			}
		}
	}
}

// cleanup performs the actual cleanup of files
func (s *CleanupService) cleanup() error {
	ctx := context.Background()

	// First, check for orphaned Redis keys (keys without corresponding files)
	pattern := rediskeys.GetLastRequestKey("*")
	iter := s.redis.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		key := iter.Val()
		filePath := rediskeys.GetFilePathFromKey(key)
		if filePath == "" {
			continue
		}

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// File doesn't exist, delete the Redis key
			if err := s.redis.Del(ctx, key).Err(); err != nil {
				log.Printf("Error deleting orphaned Redis key %s: %v", key, err)
			} else {
				log.Printf("Deleted orphaned Redis key: %s", key)
			}
		}
	}

	if err := iter.Err(); err != nil {
		log.Printf("Error scanning Redis keys: %v", err)
	}

	// Then check for files without Redis keys
	files, err := os.ReadDir(s.outputDir)
	if err != nil {
		return fmt.Errorf("failed to read output directory: %v", err)
	}

	for _, file := range files {
		filePath := filepath.Join(s.outputDir, file.Name())
		_, err := file.Info()
		if err != nil {
			log.Printf("Error getting file info for %s: %v", filePath, err)
			continue
		}

		// Skip if not a video file
		if !isVideoFile(file.Name()) {
			continue
		}

		// Check if file has been requested in the last hour
		key := rediskeys.GetLastRequestKey(filePath)
		exists, err := s.redis.Exists(ctx, key).Result()
		if err != nil {
			log.Printf("Error checking last request key for %s: %v", filePath, err)
			continue
		}

		// If key doesn't exist, file hasn't been requested in the last hour
		if exists == 0 {
			if err := s.deleteFile(filePath); err != nil {
				log.Printf("Error deleting file %s: %v", filePath, err)
			}
		}
	}

	return nil
}

// deleteFile deletes a file and its associated Redis key
func (s *CleanupService) deleteFile(filePath string) error {
	ctx := context.Background()

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %v", err)
	}

	// Delete the last request key
	key := rediskeys.GetLastRequestKey(filePath)
	if err := s.redis.Del(ctx, key).Err(); err != nil {
		log.Printf("Error deleting Redis key %s: %v", key, err)
	}

	log.Printf("Successfully deleted file and associated data: %s", filePath)
	return nil
}

// isVideoFile checks if the file is a video file
func isVideoFile(filename string) bool {
	extensions := []string{".mp4", ".webm", ".mkv"}
	for _, ext := range extensions {
		if filepath.Ext(filename) == ext {
			return true
		}
	}
	return false
}

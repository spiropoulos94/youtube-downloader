package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"spiropoulos94/youtube-downloader/internal/services"
	"strings"
	"time"

	"github.com/hibiken/asynq"
)

const TypeVideoDownload = "video:download"

// TaskStatus represents the current state of a video download task
type TaskStatus string

// Task status constants
const (
	TaskStatusPending    TaskStatus = "pending"    // Task is waiting to be processed
	TaskStatusProcessing TaskStatus = "processing" // Task is being processed
	TaskStatusCompleted  TaskStatus = "completed"  // Task has been completed successfully
	TaskStatusFailed     TaskStatus = "failed"     // Task failed to complete
)

type VideoDownloadPayload struct {
	URL      string     `json:"url"`
	FilePath string     `json:"file_path,omitempty"`
	Status   TaskStatus `json:"status"`
	Error    string     `json:"error,omitempty"`
}

func NewVideoDownloadTask(url string) (*asynq.Task, error) {
	payload := VideoDownloadPayload{
		URL:    url,
		Status: TaskStatusPending,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	task := asynq.NewTask(TypeVideoDownload, data)

	return task, nil
}

type VideoDownloadProcessor struct {
	youtubeService *services.YouTubeService
}

func NewVideoDownloadProcessor(youtubeService *services.YouTubeService) *VideoDownloadProcessor {
	return &VideoDownloadProcessor{
		youtubeService: youtubeService,
	}
}

// isTemporaryFile checks if a file is a temporary file created by yt-dlp
func isTemporaryFile(fileName string) bool {
	tempExtensions := []string{
		".part",
		".temp",
		".webm",
		".ts",
		".m4a",
		".m4v",
		".frag",
		".ytdl",
		".f248",
	}
	for _, ext := range tempExtensions {
		if strings.HasSuffix(fileName, ext) {
			return true
		}
	}
	return false
}

// waitForFileRename waits for temporary files to be processed and returns the final file path
func waitForFileRename(ctx context.Context, tempFilePath string, urlHash string) (string, error) {
	dir := filepath.Dir(tempFilePath)
	maxWaitTime := 5 * time.Minute
	checkInterval := 500 * time.Millisecond
	startTime := time.Now()

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-ticker.C:
			// Check if we've exceeded the maximum wait time
			if time.Since(startTime) > maxWaitTime {
				return "", fmt.Errorf("timeout waiting for file processing after %v", maxWaitTime)
			}

			// Look for both temporary files and the final MP4
			files, err := os.ReadDir(dir)
			if err != nil {
				log.Printf("Error reading directory: %v", err)
				continue
			}

			var tempFiles []string
			var finalFile string

			// First pass: collect all relevant files
			for _, file := range files {
				fileName := file.Name()
				if strings.Contains(fileName, urlHash) {
					if strings.HasSuffix(fileName, ".mp4") && !isTemporaryFile(fileName) {
						finalFile = fileName
						break
					} else if isTemporaryFile(fileName) {
						tempFiles = append(tempFiles, fileName)
					}
				}
			}

			// If we found the final file, verify it and return
			if finalFile != "" {
				finalPath := filepath.Join(dir, finalFile)
				if info, err := os.Stat(finalPath); err == nil && info.Size() > 0 {
					// Wait a short time to ensure the file is fully written
					time.Sleep(1 * time.Second)
					return finalPath, nil
				}
			}

			// If no temporary files exist and no final file, something went wrong
			if len(tempFiles) == 0 && finalFile == "" {
				// Only error if we've waited a reasonable amount of time
				if time.Since(startTime) > 10*time.Second {
					return "", fmt.Errorf("no temporary or final files found for hash %s", urlHash)
				}
			}
		}
	}
}

func (processor *VideoDownloadProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p VideoDownloadPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		log.Printf("Error unmarshaling payload: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %v", err)
	}

	log.Printf("Processing task...")
	log.Printf("Task payload: %s", string(t.Payload()))

	// Update status to processing
	p.Status = TaskStatusProcessing
	data, _ := json.Marshal(p)
	if _, err := t.ResultWriter().Write(data); err != nil {
		log.Printf("Error writing processing state: %v", err)
	}

	log.Printf("Downloading video from %s...", p.URL)

	filePath, err := processor.youtubeService.DownloadVideo(p.URL)
	if err != nil {
		log.Printf("Error downloading video: %v", err)
		p.Status = TaskStatusFailed
		p.Error = err.Error()
		data, _ := json.Marshal(p)
		if _, err := t.ResultWriter().Write(data); err != nil {
			log.Printf("Error writing failed state: %v", err)
		}
		return fmt.Errorf("failed to download video: %v", err)
	}

	// If the file is a temporary file, wait for it to be processed
	if isTemporaryFile(filePath) {
		urlHash := processor.youtubeService.GetURLHash(p.URL)
		finalPath, err := waitForFileRename(ctx, filePath, urlHash)
		if err != nil {
			log.Printf("Error waiting for file processing: %v", err)
			p.Status = TaskStatusFailed
			p.Error = fmt.Sprintf("Failed to complete download: %v", err)
			data, _ := json.Marshal(p)
			if _, err := t.ResultWriter().Write(data); err != nil {
				log.Printf("Error writing failed state: %v", err)
			}
			return fmt.Errorf("failed to wait for file processing: %v", err)
		}
		filePath = finalPath
	}

	log.Printf("Successfully got video at %s", filePath)
	p.Status = TaskStatusCompleted
	p.FilePath = filePath
	data, _ = json.Marshal(p)
	if _, err := t.ResultWriter().Write(data); err != nil {
		log.Printf("Error writing completed state: %v", err)
		return err
	}

	return nil
}

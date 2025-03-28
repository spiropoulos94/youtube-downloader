package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"spiropoulos94/youtube-downloader/internal/services"

	"github.com/hibiken/asynq"
)

const TypeVideoDownload = "video:download"

type VideoDownloadPayload struct {
	URL      string `json:"url"`
	FilePath string `json:"file_path,omitempty"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
}

func NewVideoDownloadTask(url string) (*asynq.Task, error) {
	payload := VideoDownloadPayload{
		URL:    url,
		Status: "pending",
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

func (processor *VideoDownloadProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p VideoDownloadPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		log.Printf("Error unmarshaling payload: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %v", err)
	}

	log.Printf("Processing task...")
	log.Printf("Task payload: %s", string(t.Payload()))

	// Update status to processing
	p.Status = "processing"
	data, _ := json.Marshal(p)
	if _, err := t.ResultWriter().Write(data); err != nil {
		log.Printf("Error writing processing state: %v", err)
	}

	log.Printf("Downloading video from %s...", p.URL)
	filePath, err := processor.youtubeService.DownloadVideo(p.URL)
	if err != nil {
		log.Printf("Error downloading video: %v", err)
		p.Status = "failed"
		p.Error = err.Error()
		data, _ := json.Marshal(p)
		if _, err := t.ResultWriter().Write(data); err != nil {
			log.Printf("Error writing failed state: %v", err)
		}
		return fmt.Errorf("failed to download video: %v", err)
	}

	log.Printf("Successfully downloaded video to %s", filePath)
	p.Status = "completed"
	p.FilePath = filePath
	data, _ = json.Marshal(p)
	if _, err := t.ResultWriter().Write(data); err != nil {
		log.Printf("Error writing completed state: %v", err)
		return err
	}

	// Return nil to mark task as completed
	return nil
}

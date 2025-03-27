package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"spiropoulos94/youtube-downloader/internal/service"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const TypeVideoDownload = "video:download"

type VideoDownloadPayload struct {
	URL      string `json:"url"`
	TaskID   string `json:"task_id"`
	FilePath string `json:"file_path,omitempty"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
}

func NewVideoDownloadTask(url string) (*asynq.Task, string, error) {
	taskID := uuid.New().String()
	payload := VideoDownloadPayload{
		URL:    url,
		TaskID: taskID,
		Status: "pending",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, "", err
	}

	return asynq.NewTask(TypeVideoDownload, data), taskID, nil
}

type VideoDownloadProcessor struct {
	youtubeService *service.YouTubeService
}

func NewVideoDownloadProcessor(youtubeService *service.YouTubeService) *VideoDownloadProcessor {
	return &VideoDownloadProcessor{
		youtubeService: youtubeService,
	}
}

func (processor *VideoDownloadProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p VideoDownloadPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %v", err)
	}

	p.Status = "processing"
	data, _ := json.Marshal(p)
	t.ResultWriter().Write(data)

	filePath, err := processor.youtubeService.DownloadVideo(p.URL)
	if err != nil {
		p.Status = "failed"
		p.Error = err.Error()
		data, _ := json.Marshal(p)
		t.ResultWriter().Write(data)
		return fmt.Errorf("failed to download video: %v", err)
	}

	p.Status = "completed"
	p.FilePath = filePath
	data, _ = json.Marshal(p)
	_, err = t.ResultWriter().Write(data)
	return err
}

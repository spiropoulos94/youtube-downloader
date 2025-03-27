package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"spiropoulos94/youtube-downloader/internal/services"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

const TypeVideoDownload = "video:download"
const ResultKeyPrefix = "video:download:result:"

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
	youtubeService *services.YouTubeService
	redisClient    *redis.Client
}

func NewVideoDownloadProcessor(youtubeService *services.YouTubeService, redisClient *redis.Client) *VideoDownloadProcessor {
	return &VideoDownloadProcessor{
		youtubeService: youtubeService,
		redisClient:    redisClient,
	}
}

func (processor *VideoDownloadProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p VideoDownloadPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		log.Printf("Error unmarshaling payload: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %v", err)
	}

	log.Printf("Processing task %s...", p.TaskID)

	// Update status to processing
	p.Status = "processing"
	data, _ := json.Marshal(p)
	t.ResultWriter().Write(data)
	processor.storeResult(ctx, p)

	log.Printf("Downloading video from %s for task %s...", p.URL, p.TaskID)
	filePath, err := processor.youtubeService.DownloadVideo(p.URL)
	if err != nil {
		log.Printf("Error downloading video for task %s: %v", p.TaskID, err)
		p.Status = "failed"
		p.Error = err.Error()
		data, _ := json.Marshal(p)
		t.ResultWriter().Write(data)
		processor.storeResult(ctx, p)
		return fmt.Errorf("failed to download video: %v", err)
	}

	log.Printf("Successfully downloaded video to %s for task %s", filePath, p.TaskID)
	p.Status = "completed"
	p.FilePath = filePath
	data, _ = json.Marshal(p)
	_, err = t.ResultWriter().Write(data)
	processor.storeResult(ctx, p)
	return err
}

// Store the task result in Redis with a 1-hour expiration.
// This is necessary because Asynq automatically deletes completed tasks from its queue,
// which means calling GetTaskStatus after completion would return "Task not found".
// By storing results in Redis separately, we can still return task status and results
// to clients even after the task is completed and removed from Asynq's queue.
// The 1-hour expiration prevents accumulating stale results indefinitely.
func (processor *VideoDownloadProcessor) storeResult(ctx context.Context, payload VideoDownloadPayload) {
	resultKey := ResultKeyPrefix + payload.TaskID
	data, _ := json.Marshal(payload)
	processor.redisClient.Set(ctx, resultKey, data, time.Hour)
}

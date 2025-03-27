package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"spiropoulos94/youtube-downloader/internal/httputils"
	"spiropoulos94/youtube-downloader/internal/services"
	"spiropoulos94/youtube-downloader/internal/tasks"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

type YouTubeHandler struct {
	youtubeService *services.YouTubeService
	client         *asynq.Client
	inspector      *asynq.Inspector
	redisClient    *redis.Client
}

func NewYouTubeHandler(youtubeService *services.YouTubeService, client *asynq.Client, inspector *asynq.Inspector, redisClient *redis.Client) *YouTubeHandler {
	return &YouTubeHandler{
		youtubeService: youtubeService,
		client:         client,
		inspector:      inspector,
		redisClient:    redisClient,
	}
}

type DownloadRequest struct {
	URL string `json:"url"`
}

type DownloadResponse struct {
	TaskID string `json:"task_id"`
}

type TaskStatusResponse struct {
	Status   string `json:"status"`
	FilePath string `json:"file_path,omitempty"`
	Error    string `json:"error,omitempty"`
}

func (h *YouTubeHandler) DownloadVideo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httputils.SendError(w, httputils.ErrMethodNotAllowed)
		return
	}

	var req DownloadRequest
	if err := httputils.ParseJSON(r, &req); err != nil {
		httputils.SendError(w, httputils.ErrBadRequest)
		return
	}

	task, taskID, err := tasks.NewVideoDownloadTask(req.URL)
	if err != nil {
		httputils.SendError(w, httputils.NewError(http.StatusInternalServerError, "Failed to create task"))
		return
	}

	// Store initial task state in Redis
	payload := tasks.VideoDownloadPayload{
		URL:    req.URL,
		TaskID: taskID,
		Status: "pending",
	}
	data, err := json.Marshal(payload)
	if err != nil {
		httputils.SendError(w, httputils.NewError(http.StatusInternalServerError, "Failed to create task"))
		return
	}
	resultKey := tasks.ResultKeyPrefix + taskID
	if err := h.redisClient.Set(r.Context(), resultKey, data, time.Hour).Err(); err != nil {
		httputils.SendError(w, httputils.NewError(http.StatusInternalServerError, "Failed to store task"))
		return
	}

	_, err = h.client.Enqueue(task)
	if err != nil {
		httputils.SendError(w, httputils.NewError(http.StatusInternalServerError, "Failed to enqueue task"))
		return
	}

	httputils.SendJSON(w, http.StatusAccepted, DownloadResponse{TaskID: taskID})
}

func (h *YouTubeHandler) GetTaskStatus(w http.ResponseWriter, r *http.Request) {
	log.Printf("GetTaskStatus: Received request for task status")

	if r.Method != http.MethodGet {
		log.Printf("GetTaskStatus: Invalid method %s, expected GET", r.Method)
		httputils.SendError(w, httputils.ErrMethodNotAllowed)
		return
	}

	taskID := chi.URLParam(r, "task_id")
	if taskID == "" {
		log.Printf("GetTaskStatus: Missing task_id parameter")
		httputils.SendError(w, httputils.ErrMissingTaskID)
		return
	}

	log.Printf("GetTaskStatus: Checking status for task %s...", taskID)

	// First try to get task info from Asynq
	task, err := h.inspector.GetTaskInfo("", taskID)
	if err != nil {
		log.Printf("GetTaskStatus: Task %s not found in Asynq, checking Redis. Error: %v", taskID, err)

		// If task not found in Asynq, check Redis for stored result
		resultKey := tasks.ResultKeyPrefix + taskID
		result, err := h.redisClient.Get(r.Context(), resultKey).Result()
		if err != nil {
			log.Printf("GetTaskStatus: Task %s not found in Redis either. Error: %v", taskID, err)
			httputils.SendError(w, httputils.ErrNotFound)
			return
		}

		var payload tasks.VideoDownloadPayload
		if err := json.Unmarshal([]byte(result), &payload); err != nil {
			log.Printf("GetTaskStatus: Error unmarshaling Redis result for task %s. Error: %v", taskID, err)
			httputils.SendError(w, httputils.ErrInternalServer)
			return
		}

		log.Printf("GetTaskStatus: Found task %s in Redis with status: %s, filepath: %s, error: %s",
			taskID, payload.Status, payload.FilePath, payload.Error)
		httputils.SendJSON(w, http.StatusOK, TaskStatusResponse{
			Status:   payload.Status,
			FilePath: payload.FilePath,
			Error:    payload.Error,
		})
		return
	}

	// Task found in Asynq, get its payload
	var payload tasks.VideoDownloadPayload
	if err := json.Unmarshal(task.Payload, &payload); err != nil {
		log.Printf("GetTaskStatus: Error unmarshaling Asynq payload for task %s. Error: %v", taskID, err)
		httputils.SendError(w, httputils.ErrInternalServer)
		return
	}

	// If we found the task by Asynq ID but the custom UUID is different,
	// try to find the task by custom UUID in Redis
	if taskID != payload.TaskID {
		log.Printf("GetTaskStatus: Found task by Asynq ID %s but custom UUID is %s, checking Redis...", taskID, payload.TaskID)
		resultKey := tasks.ResultKeyPrefix + taskID
		result, err := h.redisClient.Get(r.Context(), resultKey).Result()
		if err == nil {
			// Found in Redis by custom UUID
			if err := json.Unmarshal([]byte(result), &payload); err != nil {
				log.Printf("GetTaskStatus: Error unmarshaling Redis result for task %s. Error: %v", taskID, err)
				httputils.SendError(w, httputils.ErrInternalServer)
				return
			}
			log.Printf("GetTaskStatus: Found task %s in Redis with status: %s, filepath: %s, error: %s",
				taskID, payload.Status, payload.FilePath, payload.Error)
			httputils.SendJSON(w, http.StatusOK, TaskStatusResponse{
				Status:   payload.Status,
				FilePath: payload.FilePath,
				Error:    payload.Error,
			})
			return
		}
	}

	log.Printf("GetTaskStatus: Found task %s in Asynq with status: %s, filepath: %s, error: %s",
		taskID, payload.Status, payload.FilePath, payload.Error)
	httputils.SendJSON(w, http.StatusOK, TaskStatusResponse{
		Status:   payload.Status,
		FilePath: payload.FilePath,
		Error:    payload.Error,
	})
}

func (h *YouTubeHandler) ServeVideo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.SendError(w, httputils.ErrMethodNotAllowed)
		return
	}

	taskID := chi.URLParam(r, "task_id")
	if taskID == "" {
		httputils.SendError(w, httputils.ErrMissingTaskID)
		return
	}

	// Get task info from Redis
	resultKey := tasks.ResultKeyPrefix + taskID
	result, err := h.redisClient.Get(r.Context(), resultKey).Result()
	if err != nil {
		httputils.SendError(w, httputils.ErrNotFound)
		return
	}

	var payload tasks.VideoDownloadPayload
	if err := json.Unmarshal([]byte(result), &payload); err != nil {
		httputils.SendError(w, httputils.ErrInternalServer)
		return
	}

	if payload.Status != "completed" || payload.FilePath == "" {
		httputils.SendError(w, httputils.NewError(http.StatusBadRequest, "Video download not completed"))
		return
	}

	// Set headers for file download
	w.Header().Set("Content-Disposition", "attachment; filename=video.mp4")
	w.Header().Set("Content-Type", "video/mp4")

	// Serve the file
	http.ServeFile(w, r, payload.FilePath)
}

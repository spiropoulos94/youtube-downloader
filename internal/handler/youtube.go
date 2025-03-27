package handler

import (
	"encoding/json"
	"net/http"
	"spiropoulos94/youtube-downloader/internal/httputils"
	"spiropoulos94/youtube-downloader/internal/service"
	"spiropoulos94/youtube-downloader/internal/tasks"

	"github.com/go-chi/chi/v5"
	"github.com/hibiken/asynq"
)

type YouTubeHandler struct {
	youtubeService *service.YouTubeService
	taskClient     *asynq.Client
	inspector      *asynq.Inspector
}

func NewYouTubeHandler(youtubeService *service.YouTubeService, taskClient *asynq.Client, inspector *asynq.Inspector) *YouTubeHandler {
	return &YouTubeHandler{
		youtubeService: youtubeService,
		taskClient:     taskClient,
		inspector:      inspector,
	}
}

type DownloadRequest struct {
	URL string `json:"url"`
}

type DownloadResponse struct {
	TaskID string `json:"task_id"`
}

func (h *YouTubeHandler) DownloadVideo(w http.ResponseWriter, r *http.Request) {
	var req DownloadRequest
	if err := httputils.ParseJSON(r, &req); err != nil {
		httputils.SendError(w, httputils.ErrBadRequest)
		return
	}

	if req.URL == "" {
		httputils.SendError(w, httputils.ErrBadRequest)
		return
	}

	// Create async task
	task, taskID, err := tasks.NewVideoDownloadTask(req.URL)
	if err != nil {
		httputils.SendError(w, httputils.NewError(http.StatusInternalServerError, "Failed to create download task"))
		return
	}

	// Enqueue the task
	if _, err := h.taskClient.Enqueue(task); err != nil {
		httputils.SendError(w, httputils.NewError(http.StatusInternalServerError, "Failed to enqueue download task"))
		return
	}

	// Return task ID to client
	httputils.SendJSON(w, http.StatusAccepted, DownloadResponse{
		TaskID: taskID,
	})
}

type TaskStatusResponse struct {
	Status   string `json:"status"`
	FilePath string `json:"file_path,omitempty"`
	Error    string `json:"error,omitempty"`
}

func (h *YouTubeHandler) GetTaskStatus(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskID")
	if taskID == "" {
		httputils.SendError(w, httputils.ErrBadRequest)
		return
	}

	// Get task info from Redis
	task, err := h.inspector.GetTaskInfo("default", taskID)
	if err != nil {
		httputils.SendError(w, httputils.NewError(http.StatusNotFound, "Task not found"))
		return
	}

	// Parse task payload
	var payload tasks.VideoDownloadPayload
	if err := json.Unmarshal(task.Payload, &payload); err != nil {
		httputils.SendError(w, httputils.NewError(http.StatusInternalServerError, "Failed to parse task payload"))
		return
	}

	// Return task status
	httputils.SendJSON(w, http.StatusOK, TaskStatusResponse{
		Status:   payload.Status,
		FilePath: payload.FilePath,
		Error:    payload.Error,
	})
}

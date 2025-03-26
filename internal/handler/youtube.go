package handler

import (
	"net/http"
	"spiropoulos94/youtube-downloader/internal/httputils"
	"spiropoulos94/youtube-downloader/internal/service"
	"spiropoulos94/youtube-downloader/internal/tasks"

	"github.com/hibiken/asynq"
)

type YouTubeHandler struct {
	youtubeService *service.YouTubeService
	taskClient     *asynq.Client
}

func NewYouTubeHandler(youtubeService *service.YouTubeService, taskClient *asynq.Client) *YouTubeHandler {
	return &YouTubeHandler{
		youtubeService: youtubeService,
		taskClient:     taskClient,
	}
}

type DownloadRequest struct {
	URL     string `json:"url"`
	Quality string `json:"quality"`
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

	if req.Quality == "" {
		req.Quality = "best"
	}

	// Create async task
	task, taskID, err := tasks.NewVideoDownloadTask(req.URL, req.Quality)
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

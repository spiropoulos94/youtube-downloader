package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"spiropoulos94/youtube-downloader/internal/config"
	"spiropoulos94/youtube-downloader/internal/httputils"
	"spiropoulos94/youtube-downloader/internal/services"
	"spiropoulos94/youtube-downloader/internal/tasks"
	"spiropoulos94/youtube-downloader/internal/validators"

	"github.com/go-chi/chi/v5"
	"github.com/hibiken/asynq"
)

// YouTubeHandler implements YouTubeHandlerInterface
type YouTubeHandler struct {
	config         *config.Config
	youtubeService services.YouTubeServiceInterface
	asynqClient    *asynq.Client
	asynqInspector *asynq.Inspector
	urlValidator   validators.URLValidatorInterface
}

// NewYouTubeHandler creates a new instance of YouTubeHandler
func NewYouTubeHandler(
	config *config.Config,
	youtubeService services.YouTubeServiceInterface,
	asynqClient *asynq.Client,
	asynqInspector *asynq.Inspector,
	urlValidator validators.URLValidatorInterface,
) YouTubeHandlerInterface {
	return &YouTubeHandler{
		config:         config,
		youtubeService: youtubeService,
		asynqClient:    asynqClient,
		asynqInspector: asynqInspector,
		urlValidator:   urlValidator,
	}
}

type DownloadRequest struct {
	URL string `json:"url"`
}

type DownloadResponse struct {
	TaskID string `json:"task_id"`
}

type TaskStatusResponse struct {
	Status       tasks.TaskStatus `json:"status"`
	FilePath     string           `json:"file_path,omitempty"`
	DownloadURL  string           `json:"download_url,omitempty"`
	Error        string           `json:"error,omitempty"`
	Title        string           `json:"title,omitempty"`
	ThumbnailURL string           `json:"thumbnail_url,omitempty"`
	Duration     string           `json:"duration,omitempty"`
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

	// Validate YouTube URL using the validator
	if err := h.urlValidator.Validate(req.URL); err != nil {
		httputils.SendError(w, httputils.NewError(http.StatusBadRequest, err.Error()))
		return
	}

	task, err := tasks.NewVideoDownloadTask(req.URL)
	if err != nil {
		httputils.SendError(w, httputils.NewError(http.StatusInternalServerError, "Failed to create task"))
		return
	}

	// keep task in queue using the configured retention time
	info, err := h.asynqClient.Enqueue(task, asynq.Retention(h.config.TaskRetention))
	if err != nil {
		httputils.SendError(w, httputils.NewError(http.StatusInternalServerError, "Failed to enqueue task"))
		return
	}

	log.Printf("Task enqueued: ID=%s, URL=%s, Retention: %s", info.ID, req.URL, h.config.TaskRetention)
	httputils.SendJSON(w, http.StatusAccepted, DownloadResponse{TaskID: info.ID})
}

func (h *YouTubeHandler) GetTaskStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputils.SendError(w, httputils.ErrMethodNotAllowed)
		return
	}

	taskID := chi.URLParam(r, "task_id")
	if taskID == "" {
		httputils.SendError(w, httputils.ErrMissingTaskID)
		return
	}

	info, err := h.asynqInspector.GetTaskInfo("default", taskID)
	if err != nil {
		log.Printf("Task not found: ID=%s", taskID)
		httputils.SendError(w, httputils.ErrNotFound)
		return
	}

	var payload tasks.VideoDownloadPayload
	if err := json.Unmarshal(info.Result, &payload); err != nil {
		log.Printf("Failed to parse task result: ID=%s, Error=%v", taskID, err)
		httputils.SendError(w, httputils.ErrInternalServer)
		return
	}

	response := TaskStatusResponse{
		Status:       payload.Status,
		FilePath:     payload.FilePath,
		Error:        payload.Error,
		Title:        payload.Title,
		ThumbnailURL: payload.ThumbnailURL,
		Duration:     payload.Duration,
	}

	// Add download URL if the task is completed and we have a file path
	if payload.Status == tasks.TaskStatusCompleted && payload.FilePath != "" {
		// Use the configured BaseURL if available
		if h.config.BaseURL != "" {
			response.DownloadURL = fmt.Sprintf("%s/api/videos/%s", h.config.BaseURL, taskID)
		} else {
			// Fallback: Construct the download URL using the host from the request
			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}
			host := r.Host
			response.DownloadURL = fmt.Sprintf("%s://%s/api/videos/%s", scheme, host, taskID)
		}
	}

	httputils.SendJSON(w, http.StatusOK, response)
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

	info, err := h.asynqInspector.GetTaskInfo("default", taskID)
	if err != nil {
		log.Printf("Task not found: ID=%s", taskID)
		httputils.SendError(w, httputils.ErrNotFound)
		return
	}

	var payload tasks.VideoDownloadPayload
	if err := json.Unmarshal(info.Result, &payload); err != nil {
		log.Printf("Failed to parse task result: ID=%s, Error=%v", taskID, err)
		httputils.SendError(w, httputils.ErrInternalServer)
		return
	}

	if payload.Status != tasks.TaskStatusCompleted || payload.FilePath == "" {
		log.Printf("Video not ready: ID=%s, Status=%s", taskID, payload.Status)
		httputils.SendError(w, httputils.NewError(http.StatusBadRequest, "Video download not completed"))
		return
	}

	// Open the file before setting headers
	file, err := os.Open(payload.FilePath)
	if err != nil {
		log.Printf("Failed to open file: ID=%s, Error=%v", taskID, err)
		httputils.SendError(w, httputils.ErrInternalServer)
		return
	}
	defer file.Close()

	// Get file info for content length
	fileInfo, err := file.Stat()
	if err != nil {
		log.Printf("Failed to get file info: ID=%s, Error=%v", taskID, err)
		httputils.SendError(w, httputils.ErrInternalServer)
		return
	}

	// Extract the original filename from the path and escape it for Content-Disposition header
	downloadFileName := h.youtubeService.GetOriginalFilename(payload.FilePath, true)

	// Set headers with the original title for the download
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, downloadFileName))
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	log.Printf("Serving video: ID=%s, File=%s, Title=%s", taskID, payload.FilePath, downloadFileName)

	// Serve the file
	http.ServeContent(w, r, fileInfo.Name(), fileInfo.ModTime(), file)
}

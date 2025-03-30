package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"spiropoulos94/youtube-downloader/internal/httputils"
	"spiropoulos94/youtube-downloader/internal/services"
	"spiropoulos94/youtube-downloader/internal/tasks"
	"spiropoulos94/youtube-downloader/internal/validators"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hibiken/asynq"
)

type YouTubeHandler struct {
	youtubeService *services.YouTubeService
	client         *asynq.Client
	inspector      *asynq.Inspector
	urlValidator   *validators.YouTubeURLValidator
}

func NewYouTubeHandler(youtubeService *services.YouTubeService, client *asynq.Client, inspector *asynq.Inspector) *YouTubeHandler {
	return &YouTubeHandler{
		youtubeService: youtubeService,
		client:         client,
		inspector:      inspector,
		urlValidator:   validators.NewYouTubeURLValidator(),
	}
}

type DownloadRequest struct {
	URL string `json:"url"`
}

type DownloadResponse struct {
	TaskID string `json:"task_id"`
}

type TaskStatusResponse struct {
	Status   tasks.TaskStatus `json:"status"`
	FilePath string           `json:"file_path,omitempty"`
	Error    string           `json:"error,omitempty"`
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

	// keep task in queue for 1 hour so that it can be accessed even after its completed
	info, err := h.client.Enqueue(task, asynq.Retention(1*time.Hour))
	if err != nil {
		httputils.SendError(w, httputils.NewError(http.StatusInternalServerError, "Failed to enqueue task"))
		return
	}

	log.Printf("Task enqueued: ID=%s, URL=%s", info.ID, req.URL)
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

	info, err := h.inspector.GetTaskInfo("default", taskID)
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
		Status:   payload.Status,
		FilePath: payload.FilePath,
		Error:    payload.Error,
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

	info, err := h.inspector.GetTaskInfo("default", taskID)
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

	// Set headers
	w.Header().Set("Content-Disposition", "attachment; filename=video.mp4")
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	log.Printf("Serving video: ID=%s, File=%s", taskID, payload.FilePath)

	// Create cleanup function
	cleanup := func() {
		if err := h.youtubeService.DecrementRefCount(payload.FilePath); err != nil {
			log.Printf("Failed to decrement reference count: ID=%s, Error=%v", taskID, err)
		}
	}

	// Create cleanup reader that will run the cleanup function when the reader is closed aka when the file is downloaded
	cleanupReader := httputils.NewCleanupReader(file, cleanup)

	// Serve the file
	http.ServeContent(w, r, fileInfo.Name(), fileInfo.ModTime(), cleanupReader)
}

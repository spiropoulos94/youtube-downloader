package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"spiropoulos94/youtube-downloader/internal/service"
)

type YouTubeHandler struct {
	youtubeService *service.YouTubeService
}

func NewYouTubeHandler(youtubeService *service.YouTubeService) *YouTubeHandler {
	return &YouTubeHandler{
		youtubeService: youtubeService,
	}
}

type DownloadRequest struct {
	URL     string `json:"url"`
	Quality string `json:"quality"`
}

func (h *YouTubeHandler) DownloadVideo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	if req.Quality == "" {
		req.Quality = "best"
	}

	// Download the video
	filePath, err := h.youtubeService.DownloadVideo(req.URL, req.Quality)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Failed to open video file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Get file info for content length
	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "Failed to get file info", http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(filePath)))

	// Stream the file to the response
	http.ServeFile(w, r, filePath)
}

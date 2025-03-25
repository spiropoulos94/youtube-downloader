package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"spiropoulos94/youtube-downloader/internal/httputils"
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

type DownloadResponse struct {
	FilePath string `json:"file_path"`
}

func (h *YouTubeHandler) DownloadVideo(w http.ResponseWriter, r *http.Request) {
	var req DownloadRequest
	if err := httputils.ParseJSON(r, &req); err != nil {
		httputils.SendError(w, httputils.ErrBadRequest)
		return
	}

	if req.URL == "" {
		httputils.SendError(w, httputils.NewError(http.StatusBadRequest, "URL is required"))
		return
	}

	if req.Quality == "" {
		req.Quality = "best"
	}

	// Download the video
	filePath, err := h.youtubeService.DownloadVideo(req.URL, req.Quality)
	if err != nil {
		httputils.SendError(w, err)
		return
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		httputils.SendError(w, httputils.NewError(http.StatusInternalServerError, "Failed to open video file"))
		return
	}
	defer file.Close()

	// Get file info for content length
	fileInfo, err := file.Stat()
	if err != nil {
		httputils.SendError(w, httputils.NewError(http.StatusInternalServerError, "Failed to get file info"))
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(filePath)))

	// Stream the file to the response
	http.ServeFile(w, r, filePath)
}

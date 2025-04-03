package handlers

import (
	"net/http"
)

// YouTubeHandlerInterface defines the contract for YouTube-related HTTP handlers
type YouTubeHandlerInterface interface {
	DownloadVideo(w http.ResponseWriter, r *http.Request)
	GetTaskStatus(w http.ResponseWriter, r *http.Request)
	ServeVideo(w http.ResponseWriter, r *http.Request)
}

// FrontendHandlerInterface defines the contract for frontend-related HTTP handlers
type FrontendHandlerInterface interface {
	ServeFrontend(w http.ResponseWriter, r *http.Request)
}

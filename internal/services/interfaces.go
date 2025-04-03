package services

import (
	"net/http"
)

// YouTubeServiceInterface defines the contract for YouTube-related operations
type YouTubeServiceInterface interface {
	UpdateLastRequestTime(filePath string) error
	GetURLHash(url string) string
	DownloadVideo(url string) (*VideoData, error)
	StoreMetadata(filePath string, metadata *VideoMetadata) error
	GetStoredMetadata(filePath string) (*VideoMetadata, error)
	GetOriginalFilename(filePath string, escape bool) string
}

// FrontendServiceInterface defines the contract for frontend-related operations
type FrontendServiceInterface interface {
	ServeStaticFiles(w http.ResponseWriter, r *http.Request)
}

// CleanupServiceInterface defines the contract for cleanup operations
type CleanupServiceInterface interface {
	Start()
	Stop()
}

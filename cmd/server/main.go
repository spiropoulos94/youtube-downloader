package main

import (
	"log"
	"net/http"
	"spiropoulos94/youtube-downloader/internal/container"
)

func main() {
	// Initialize container with command line flags
	container, err := container.InitContainer()
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}

	// Get handler from container
	youtubeHandler := container.GetYouTubeHandler()

	// Set up routes
	http.HandleFunc("/download", youtubeHandler.DownloadVideo)

	// Start server
	log.Printf("Server starting on port %s...", container.GetPort())
	if err := http.ListenAndServe(":"+container.GetPort(), nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

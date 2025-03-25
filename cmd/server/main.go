package main

import (
	"log"
	"spiropoulos94/youtube-downloader/internal/container"
)

func main() {
	// Initialize container with command line flags
	container, err := container.InitContainer()
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}

	// Start server
	log.Printf("Server starting on port %s...", container.GetPort())
	if err := container.StartServer(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"spiropoulos94/youtube-downloader/internal/container"
)

func main() {
	// Initialize container with command line flags
	container, err := container.InitContainer()
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}

	// Create a channel to listen for errors coming from the server
	serverErrors := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s...", container.GetPort())
		serverErrors <- container.StartServer()
	}()

	// Create a channel to listen for OS signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking select waiting for either a server error or a shutdown signal
	select {
	case err := <-serverErrors:
		log.Printf("Server error: %v", err)

	case sig := <-shutdown:
		log.Printf("Received signal %v, initiating shutdown...", sig)

		// Close the container (this will stop workers and close Redis connection)
		if err := container.Close(); err != nil {
			log.Printf("Error during container shutdown: %v", err)
		}

		log.Println("Shutdown complete")
	}
}

package main

import (
	"flag"
	"fmt"
	"log"
	"spiropoulos94/youtube-downloader/internal/services"
)

func main() {
	// Define command line flags
	url := flag.String("url", "", "YouTube video URL")
	outputDir := flag.String("output", "downloads", "Output directory for downloaded videos")
	flag.Parse()

	// Validate URL
	if *url == "" {
		fmt.Println("Please provide a YouTube URL using the -url flag")
		fmt.Println("Usage example:")
		fmt.Println("  ./cli -url https://www.youtube.com/watch?v=... -output downloads")
		return
	}

	// Create YouTube service
	youtubeService := services.NewYouTubeService(*outputDir, nil) // we don't need redis client for the cli, since we are not using the server and the video is instantly downloaded and accessible to the user

	// Download video
	fmt.Printf("Downloading video from: %s\n", *url)
	filePath, err := youtubeService.DownloadVideo(*url)
	if err != nil {
		log.Fatalf("Failed to download video: %v", err)
	}

	fmt.Printf("Download completed! Video saved to: %s\n", filePath)
}

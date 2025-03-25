package main

import (
	"flag"
	"fmt"
	"log"
	"spiropoulos94/youtube-downloader/internal/service"
)

func main() {
	// Define command line flags
	url := flag.String("url", "", "YouTube video URL")
	outputDir := flag.String("output", "downloads", "Output directory for downloaded videos")
	quality := flag.String("quality", "best", "Video quality (best, 1080p, 720p, etc.)")
	flag.Parse()

	// Validate URL
	if *url == "" {
		fmt.Println("Please provide a YouTube URL using the -url flag")
		fmt.Println("Usage example:")
		fmt.Println("  ./cli -url https://www.youtube.com/watch?v=... -quality 1080p -output downloads")
		return
	}

	// Create YouTube service
	youtubeService := service.NewYouTubeService(*outputDir)

	// Download video
	fmt.Printf("Downloading video from: %s\n", *url)
	filePath, err := youtubeService.DownloadVideo(*url, *quality)
	if err != nil {
		log.Fatalf("Failed to download video: %v", err)
	}

	fmt.Printf("Download completed! Video saved to: %s\n", filePath)
}

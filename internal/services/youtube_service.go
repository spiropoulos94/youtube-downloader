package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type YouTubeService struct {
	outputDir string
}

func NewYouTubeService(outputDir string) *YouTubeService {
	return &YouTubeService{
		outputDir: outputDir,
	}
}

func (s *YouTubeService) DownloadVideo(url string) (string, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(s.outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %v", err)
	}

	// Check if yt-dlp is installed
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return "", fmt.Errorf("yt-dlp is not installed. Please install it first:\nOn macOS: brew install yt-dlp\nOn Linux: sudo apt install yt-dlp or sudo pip install yt-dlp")
	}

	// Construct the yt-dlp command for download
	// Note about format selection:
	// YouTube often provides video content in two separate streams:
	// 1. A video stream (containing just the video)
	// 2. An audio stream (containing just the audio)
	//
	// By not specifying a format with -f, we let yt-dlp:
	// 1. Find the best quality video stream (e.g., 4K video)
	// 2. Find the best quality audio stream (e.g., high bitrate audio)
	// 3. Automatically download and merge these streams together
	//
	// This approach typically results in better quality than using -f best because:
	// - It can select the highest quality video stream (which might not be available in pre-merged formats)
	// - It can select the highest quality audio stream (which might not be available in pre-merged formats)
	// - It combines these best-quality streams into a single file
	//
	// For example, a video might have:
	// - Pre-merged format: 1080p video with 128kbps audio
	// - Separate streams: 4K video + 256kbps audio
	//
	// Using -f best would only see the pre-merged 1080p version,
	// while letting yt-dlp choose automatically would get the 4K video with better audio.
	cmd := exec.Command("yt-dlp",
		"-o", filepath.Join(s.outputDir, "%(title)s.%(ext)s"),
		"--merge-output-format", "mp4", // Force output to be MP4
		"--no-playlist",
		url)

	// Set up command output
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to download video: %v", err)
	}

	// Get the downloaded file path
	files, err := os.ReadDir(s.outputDir)
	if err != nil {
		return "", fmt.Errorf("failed to read output directory: %v", err)
	}

	// Find the most recently created file
	var latestFile string
	var latestTime int64
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Unix() > latestTime {
			latestTime = info.ModTime().Unix()
			latestFile = filepath.Join(s.outputDir, info.Name())
		}
	}

	if latestFile == "" {
		return "", fmt.Errorf("no downloaded file found")
	}

	return latestFile, nil
}

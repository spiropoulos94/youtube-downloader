package service

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

func (s *YouTubeService) DownloadVideo(url string, quality string) (string, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(s.outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %v", err)
	}

	// Check if yt-dlp is installed
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return "", fmt.Errorf("yt-dlp is not installed. Please install it first:\nOn macOS: brew install yt-dlp\nOn Linux: sudo apt install yt-dlp or sudo pip install yt-dlp")
	}

	// Construct the yt-dlp command
	cmd := exec.Command("yt-dlp",
		"-f", quality,
		"-o", filepath.Join(s.outputDir, "%(title)s.%(ext)s"),
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

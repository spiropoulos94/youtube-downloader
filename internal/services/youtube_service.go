package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"spiropoulos94/youtube-downloader/internal/config"
	"spiropoulos94/youtube-downloader/internal/rediskeys"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type YouTubeService struct {
	config *config.Config
	redis  *redis.Client
}

func NewYouTubeService(config *config.Config, redis *redis.Client) *YouTubeService {
	return &YouTubeService{
		config: config,
		redis:  redis,
	}
}

// UpdateLastRequestTime updates the last request time for a file
func (s *YouTubeService) UpdateLastRequestTime(filePath string) error {
	ctx := context.Background()
	key := rediskeys.GetLastRequestKey(filePath)
	now := time.Now().Format(time.RFC3339)

	if err := s.redis.Set(ctx, key, now, s.config.TaskRetention).Err(); err != nil {
		return fmt.Errorf("failed to update last request time: %v", err)
	}
	return nil
}

// GetURLHash returns a hash of the URL that can be used to identify the video
func (s *YouTubeService) GetURLHash(url string) string {
	urlHash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(urlHash[:8]) // Use first 8 chars of hash
}

// VideoData contains both file path and metadata
type VideoData struct {
	FilePath     string
	Title        string
	ThumbnailURL string
	Duration     string
}

// DownloadVideo downloads a video from YouTube and returns the file path and metadata in a single operation
func (s *YouTubeService) DownloadVideo(url string) (*VideoData, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(s.config.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %v", err)
	}

	// Check if yt-dlp is installed
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return nil, fmt.Errorf("yt-dlp is not installed. Please install it first:\nOn macOS: brew install yt-dlp\nOn Linux: sudo apt install yt-dlp or sudo pip install yt-dlp")
	}

	// Get URL hash
	urlHashStr := s.GetURLHash(url)

	// Check if video already exists
	existingFiles, err := os.ReadDir(s.config.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read output directory: %v", err)
	}

	// Look for existing video with the same URL hash
	for _, file := range existingFiles {
		// check if the file name contains the url hash of the new video url
		if strings.HasSuffix(file.Name(), urlHashStr+".mp4") {
			filePath := filepath.Join(s.config.OutputDir, file.Name())
			// Verify file exists and is readable
			if _, err := os.Stat(filePath); err == nil {
				// update last request time since the file already exists
				if err := s.UpdateLastRequestTime(filePath); err != nil {
					return nil, err
				}

				// Check if we have metadata stored in Redis
				metadata, err := s.GetStoredMetadata(filePath)
				if err != nil {
					// If no stored metadata, fetch it from youtube and store it
					log.Printf("No stored metadata found for %s, fetching...", filePath)
					metadata, err = s.fetchMetadata(url)
					if err != nil {
						log.Printf("Warning: Failed to fetch metadata for existing video: %v", err)
						// Return the file even if metadata fetch fails
						return &VideoData{
							FilePath: filePath,
						}, nil
					}

					// Store the fetched metadata in Redis for future use
					if err := s.StoreMetadata(filePath, metadata); err != nil {
						log.Printf("Warning: Failed to store metadata: %v", err)
					}
				}

				return &VideoData{
					FilePath:     filePath,
					Title:        metadata.Title,
					ThumbnailURL: metadata.ThumbnailURL,
					Duration:     metadata.Duration,
				}, nil
			}
		}
	}

	// If we need to download, get both metadata and download in one efficient operation
	// First, set up output template
	outputTemplate := filepath.Join(s.config.OutputDir, fmt.Sprintf("%%(title)s_%s.%%(ext)s", urlHashStr))

	// Combined process: first get metadata and then download
	// This is the most efficient approach that requires just one yt-dlp process
	cmd := exec.Command("yt-dlp",
		"--dump-json",        // Print JSON metadata to stdout
		"--no-simulate",      // Actually download the video
		"-o", outputTemplate, // Set output template
		"--merge-output-format", "mp4", // Force output to be MP4
		"--windows-filenames", // Only restrict characters that are illegal in Windows
		"--no-playlist",       // Don't download playlists
		"--quiet",             // Don't print progress (we'll only get the JSON)
		url)

	// Capture stdout which will contain the JSON metadata
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	// Start command
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start download: %v", err)
	}

	// Read JSON metadata from stdout
	metadataBytes, err := io.ReadAll(stdout)
	if err != nil {
		// If we fail to read metadata, don't fail the download
		log.Printf("Warning: Failed to read metadata: %v", err)
		// Let the download continue
	}

	// Parse metadata if we got it
	var metadata VideoMetadata
	if len(metadataBytes) > 0 {
		var data map[string]interface{}
		if err := json.Unmarshal(metadataBytes, &data); err != nil {
			log.Printf("Warning: Failed to parse metadata: %v", err)
		} else {
			// Extract metadata fields
			title, _ := data["title"].(string)

			// Get thumbnail URL
			var thumbnailURL string
			if thumbnails, ok := data["thumbnails"].([]interface{}); ok && len(thumbnails) > 0 {
				// Try to get a medium quality thumbnail, or use the last one as fallback
				bestThumbnail := thumbnails[len(thumbnails)-1].(map[string]interface{})
				for _, thumb := range thumbnails {
					if t, ok := thumb.(map[string]interface{}); ok {
						if res, ok := t["resolution"].(string); ok && res == "medium" {
							bestThumbnail = t
							break
						}
					}
				}
				thumbnailURL, _ = bestThumbnail["url"].(string)
			}

			// Get duration
			var duration string
			if durationSecs, ok := data["duration"].(float64); ok {
				minutes := int(durationSecs) / 60
				seconds := int(durationSecs) % 60
				duration = fmt.Sprintf("%d:%02d", minutes, seconds)
			}

			metadata = VideoMetadata{
				Title:        title,
				ThumbnailURL: thumbnailURL,
				Duration:     duration,
			}
		}
	}

	// Wait for download to complete
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("failed to download video: %v", err)
	}

	// At this point, the video has been downloaded
	// Find the newly downloaded file
	files, err := os.ReadDir(s.config.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read output directory: %v", err)
	}

	// Find the file with our URL hash
	for _, file := range files {
		if strings.HasSuffix(file.Name(), urlHashStr+".mp4") {
			filePath := filepath.Join(s.config.OutputDir, file.Name())
			// Update last request time for the newly downloaded file
			if err := s.UpdateLastRequestTime(filePath); err != nil {
				return nil, err
			}

			// Store the metadata in Redis for future use
			if err := s.StoreMetadata(filePath, &metadata); err != nil {
				log.Printf("Warning: Failed to store metadata: %v", err)
			}

			return &VideoData{
				FilePath:     filePath,
				Title:        metadata.Title,
				ThumbnailURL: metadata.ThumbnailURL,
				Duration:     metadata.Duration,
			}, nil
		}
	}

	return nil, fmt.Errorf("no downloaded file found with hash %s", urlHashStr)
}

// StoreMetadata stores video metadata in Redis
func (s *YouTubeService) StoreMetadata(filePath string, metadata *VideoMetadata) error {
	ctx := context.Background()
	key := rediskeys.GetMetadataKey(filePath)
	data, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %v", err)
	}

	if err := s.redis.Set(ctx, key, data, s.config.TaskRetention).Err(); err != nil {
		return fmt.Errorf("failed to store metadata: %v", err)
	}
	return nil
}

// GetStoredMetadata retrieves video metadata from Redis
func (s *YouTubeService) GetStoredMetadata(filePath string) (*VideoMetadata, error) {
	ctx := context.Background()
	key := rediskeys.GetMetadataKey(filePath)

	// Get metadata from Redis
	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("no metadata found")
		}
		return nil, fmt.Errorf("failed to get metadata: %v", err)
	}

	// Parse metadata
	var metadata VideoMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %v", err)
	}

	return &metadata, nil
}

// fetchMetadata is a helper method to get video metadata
func (s *YouTubeService) fetchMetadata(url string) (*VideoMetadata, error) {
	// Run yt-dlp to get video info
	cmd := exec.Command("yt-dlp",
		"--dump-json",
		"--no-playlist",
		"--skip-download",
		url)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch video metadata: %v", err)
	}

	// Parse the JSON output
	var data map[string]interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		return nil, fmt.Errorf("failed to parse video metadata: %v", err)
	}

	// Extract relevant information
	title, _ := data["title"].(string)

	// Get the best thumbnail
	var thumbnailURL string
	if thumbnails, ok := data["thumbnails"].([]interface{}); ok && len(thumbnails) > 0 {
		// Try to get a medium quality thumbnail, or use the last one as fallback
		bestThumbnail := thumbnails[len(thumbnails)-1].(map[string]interface{})
		for _, thumb := range thumbnails {
			if t, ok := thumb.(map[string]interface{}); ok {
				if res, ok := t["resolution"].(string); ok && res == "medium" {
					bestThumbnail = t
					break
				}
			}
		}
		thumbnailURL, _ = bestThumbnail["url"].(string)
	}

	// Get duration if available
	var duration string
	if durationSecs, ok := data["duration"].(float64); ok {
		minutes := int(durationSecs) / 60
		seconds := int(durationSecs) % 60
		duration = fmt.Sprintf("%d:%02d", minutes, seconds)
	}

	return &VideoMetadata{
		Title:        title,
		ThumbnailURL: thumbnailURL,
		Duration:     duration,
	}, nil
}

// GetOriginalFilename extracts the original title from a downloaded video filename.
// The format is expected to be "title_hash.mp4", and this function returns "title.mp4".
// If escape is true, it also escapes special characters for use in Content-Disposition headers.
func (s *YouTubeService) GetOriginalFilename(filePath string, escape bool) string {
	fileName := filepath.Base(filePath)
	// Remove the hash part to get just the title portion
	lastUnderscore := strings.LastIndex(fileName, "_")
	titlePart := fileName
	if lastUnderscore != -1 {
		titlePart = fileName[:lastUnderscore]
	}
	// Ensure the extension is preserved
	if !strings.HasSuffix(titlePart, ".mp4") {
		titlePart += ".mp4"
	}

	if escape {
		// Escape double quotes for Content-Disposition header
		titlePart = strings.ReplaceAll(titlePart, `"`, `\"`)
	}

	return titlePart
}

// VideoMetadata contains basic information about a YouTube video
type VideoMetadata struct {
	Title        string `json:"title"`
	ThumbnailURL string `json:"thumbnail_url"`
	Duration     string `json:"duration,omitempty"`
}

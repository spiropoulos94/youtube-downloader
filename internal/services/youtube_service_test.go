package services

import (
	"spiropoulos94/youtube-downloader/internal/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test GetURLHash
func TestGetURLHash(t *testing.T) {
	// Create a real service
	cfg := &config.Config{}
	service := &YouTubeService{
		config: cfg,
	}

	// Test cases
	tests := []struct {
		name string
		url  string
	}{
		{
			name: "Basic URL",
			url:  "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		},
		{
			name: "Empty URL",
			url:  "",
		},
		{
			name: "Different URL",
			url:  "https://www.youtube.com/watch?v=differentID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetURLHash(tt.url)

			// Ensure we got a result
			assert.NotEmpty(t, result)

			// Ensure it's 16 characters (8 bytes hex encoded)
			assert.Equal(t, 16, len(result))

			// Call again to verify consistency
			resultAgain := service.GetURLHash(tt.url)
			assert.Equal(t, result, resultAgain)
		})
	}

	// Extra test to ensure different URLs produce different hashes
	hash1 := service.GetURLHash("https://www.youtube.com/watch?v=videoA")
	hash2 := service.GetURLHash("https://www.youtube.com/watch?v=videoB")
	assert.NotEqual(t, hash1, hash2)
}

// Test GetOriginalFilename
func TestGetOriginalFilename(t *testing.T) {
	// Create a real service
	cfg := &config.Config{}
	service := &YouTubeService{
		config: cfg,
	}

	// Test cases
	tests := []struct {
		name     string
		filePath string
		escape   bool
		expected string
	}{
		{
			name:     "Basic filename without escape",
			filePath: "/tmp/videos/Test Video_abc123.mp4",
			escape:   false,
			expected: "Test Video.mp4",
		},
		{
			name:     "Basic filename with escape",
			filePath: `/tmp/videos/Test "Video"_abc123.mp4`,
			escape:   true,
			expected: `Test \"Video\".mp4`,
		},
		{
			name:     "Filename without hash",
			filePath: "/tmp/videos/Test Video.mp4",
			escape:   false,
			expected: "Test Video.mp4",
		},
		{
			name:     "Filename with multiple underscores",
			filePath: "/tmp/videos/Test_Video_Name_abc123.mp4",
			escape:   false,
			expected: "Test_Video_Name.mp4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetOriginalFilename(tt.filePath, tt.escape)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test VideoMetadata
func TestVideoMetadata(t *testing.T) {
	// Test that the struct can be created and fields accessed
	metadata := &VideoMetadata{
		Title:        "Test Video",
		ThumbnailURL: "https://example.com/thumbnail.jpg",
		Duration:     "3:45",
	}

	assert.Equal(t, "Test Video", metadata.Title)
	assert.Equal(t, "https://example.com/thumbnail.jpg", metadata.ThumbnailURL)
	assert.Equal(t, "3:45", metadata.Duration)
}

// Test VideoData
func TestVideoData(t *testing.T) {
	// Test that the struct can be created and fields accessed
	videoData := &VideoData{
		FilePath:     "/path/to/video.mp4",
		Title:        "Test Video",
		ThumbnailURL: "https://example.com/thumbnail.jpg",
		Duration:     "3:45",
	}

	assert.Equal(t, "/path/to/video.mp4", videoData.FilePath)
	assert.Equal(t, "Test Video", videoData.Title)
	assert.Equal(t, "https://example.com/thumbnail.jpg", videoData.ThumbnailURL)
	assert.Equal(t, "3:45", videoData.Duration)
}

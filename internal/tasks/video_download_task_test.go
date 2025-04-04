package tasks

import (
	"encoding/json"
	"spiropoulos94/youtube-downloader/internal/services"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Create a mock implementation of YouTubeServiceInterface for testing
type mockYouTubeService struct {
	services.YouTubeServiceInterface
}

func TestNewVideoDownloadTask(t *testing.T) {
	testURL := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"

	// Create task
	task, err := NewVideoDownloadTask(testURL)

	// Assert no error occurred
	require.NoError(t, err)

	// Assert task type is correct
	assert.Equal(t, TypeVideoDownload, task.Type())

	// Decode payload and verify it contains the expected data
	var payload VideoDownloadPayload
	err = json.Unmarshal(task.Payload(), &payload)
	require.NoError(t, err)

	// Verify payload fields
	assert.Equal(t, testURL, payload.URL)
	assert.Equal(t, TaskStatusPending, payload.Status)
	assert.Empty(t, payload.FilePath)
	assert.Empty(t, payload.Error)
}

func TestNewVideoDownloadProcessor(t *testing.T) {
	// Create a mock YouTube service
	mockService := &mockYouTubeService{}

	// Create processor
	processor := NewVideoDownloadProcessor(mockService)

	// Assert processor is properly initialized
	assert.NotNil(t, processor)
	assert.Equal(t, mockService, processor.youtubeService)
}

func TestIsTemporaryFile(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		expected bool
	}{
		{
			name:     "Regular MP4 file",
			fileName: "video.mp4",
			expected: false,
		},
		{
			name:     "Part file",
			fileName: "video.part",
			expected: true,
		},
		{
			name:     "Temp file",
			fileName: "video.temp",
			expected: true,
		},
		{
			name:     "WebM file",
			fileName: "video.webm",
			expected: true,
		},
		{
			name:     "TS file",
			fileName: "video.ts",
			expected: true,
		},
		{
			name:     "M4A file",
			fileName: "video.m4a",
			expected: true,
		},
		{
			name:     "M4V file",
			fileName: "video.m4v",
			expected: true,
		},
		{
			name:     "Frag file",
			fileName: "video.frag",
			expected: true,
		},
		{
			name:     "YTDL file",
			fileName: "video.ytdl",
			expected: true,
		},
		{
			name:     "F248 file",
			fileName: "video.f248",
			expected: true,
		},
		{
			name:     "File with temporary extension in the middle",
			fileName: "video.part.mp4",
			expected: false,
		},
		{
			name:     "File with regular extension in path",
			fileName: "/tmp/downloads/video.mp4",
			expected: false,
		},
		{
			name:     "File with temporary extension in path",
			fileName: "/tmp/downloads/video.part",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTemporaryFile(tt.fileName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTaskStatusConstants(t *testing.T) {
	// Verify task status constants
	assert.Equal(t, TaskStatus("pending"), TaskStatusPending)
	assert.Equal(t, TaskStatus("processing"), TaskStatusProcessing)
	assert.Equal(t, TaskStatus("completed"), TaskStatusCompleted)
	assert.Equal(t, TaskStatus("failed"), TaskStatusFailed)

	// Verify they are distinct
	statuses := map[TaskStatus]bool{
		TaskStatusPending:    true,
		TaskStatusProcessing: true,
		TaskStatusCompleted:  true,
		TaskStatusFailed:     true,
	}
	assert.Len(t, statuses, 4, "All task statuses should be distinct")
}

func TestVideoDownloadPayloadStructure(t *testing.T) {
	// Create a sample payload with all fields set
	payload := VideoDownloadPayload{
		URL:          "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		FilePath:     "/path/to/video.mp4",
		Status:       TaskStatusCompleted,
		Error:        "sample error",
		Title:        "Sample Video",
		ThumbnailURL: "https://img.youtube.com/vi/dQw4w9WgXcQ/default.jpg",
		Duration:     "3:32",
	}

	// Marshal to JSON
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	// Unmarshal back to verify field mapping
	var decoded VideoDownloadPayload
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify all fields match
	assert.Equal(t, payload.URL, decoded.URL)
	assert.Equal(t, payload.FilePath, decoded.FilePath)
	assert.Equal(t, payload.Status, decoded.Status)
	assert.Equal(t, payload.Error, decoded.Error)
	assert.Equal(t, payload.Title, decoded.Title)
	assert.Equal(t, payload.ThumbnailURL, decoded.ThumbnailURL)
	assert.Equal(t, payload.Duration, decoded.Duration)

	// Check JSON field names using a map
	var jsonMap map[string]interface{}
	err = json.Unmarshal(data, &jsonMap)
	require.NoError(t, err)

	assert.Contains(t, jsonMap, "url")
	assert.Contains(t, jsonMap, "file_path")
	assert.Contains(t, jsonMap, "status")
	assert.Contains(t, jsonMap, "error")
	assert.Contains(t, jsonMap, "title")
	assert.Contains(t, jsonMap, "thumbnail_url")
	assert.Contains(t, jsonMap, "duration")
}

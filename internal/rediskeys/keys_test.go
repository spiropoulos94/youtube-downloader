package rediskeys

import (
	"testing"
)

func TestGetLastRequestKey(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{
			name:     "regular file path",
			filePath: "videos/example.mp4",
			expected: "video:lastrequest:videos/example.mp4",
		},
		{
			name:     "empty file path",
			filePath: "",
			expected: "video:lastrequest:",
		},
		{
			name:     "file path with special characters",
			filePath: "videos/my-video (2023) [1080p].mp4",
			expected: "video:lastrequest:videos/my-video (2023) [1080p].mp4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetLastRequestKey(tt.filePath)
			if result != tt.expected {
				t.Errorf("GetLastRequestKey(%q) = %q, want %q", tt.filePath, result, tt.expected)
			}
		})
	}
}

func TestGetFilePathFromKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "valid key",
			key:      "video:lastrequest:videos/example.mp4",
			expected: "videos/example.mp4",
		},
		{
			name:     "invalid prefix",
			key:      "something:else:videos/example.mp4",
			expected: "",
		},
		{
			name:     "exact prefix only",
			key:      "video:lastrequest:",
			expected: "",
		},
		{
			name:     "empty key",
			key:      "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFilePathFromKey(tt.key)
			if result != tt.expected {
				t.Errorf("GetFilePathFromKey(%q) = %q, want %q", tt.key, result, tt.expected)
			}
		})
	}
}

func TestGetMetadataKey(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{
			name:     "regular file path",
			filePath: "videos/example.mp4",
			expected: "video:metadata:videos/example.mp4",
		},
		{
			name:     "empty file path",
			filePath: "",
			expected: "video:metadata:",
		},
		{
			name:     "file path with special characters",
			filePath: "videos/my-video (2023) [1080p].mp4",
			expected: "video:metadata:videos/my-video (2023) [1080p].mp4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMetadataKey(tt.filePath)
			if result != tt.expected {
				t.Errorf("GetMetadataKey(%q) = %q, want %q", tt.filePath, result, tt.expected)
			}
		})
	}
}

func TestKeyRoundTrip(t *testing.T) {
	// Test that we can get a file path back from a key generated from that same path
	filePaths := []string{
		"videos/example.mp4",
		"path/to/some/nested/file.mp4",
		"video with spaces.mp4",
		"special_chars!@#$%^&*().mp4",
	}

	for _, path := range filePaths {
		key := GetLastRequestKey(path)
		extractedPath := GetFilePathFromKey(key)

		if extractedPath != path {
			t.Errorf("Round trip failed: original=%q, key=%q, extracted=%q",
				path, key, extractedPath)
		}
	}
}

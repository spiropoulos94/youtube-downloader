package rediskeys

import (
	"fmt"
	"strings"
)

// GetLastRequestKey returns the Redis key for a file's last request time
func GetLastRequestKey(filePath string) string {
	return fmt.Sprintf("video:lastrequest:%s", filePath)
}

// GetFilePathFromKey extracts the file path from a Redis key
func GetFilePathFromKey(key string) string {
	prefix := "video:lastrequest:"
	if !strings.HasPrefix(key, prefix) {
		return ""
	}
	return strings.TrimPrefix(key, prefix)
}

// GetMetadataKey returns the Redis key for a video's metadata
func GetMetadataKey(filePath string) string {
	return fmt.Sprintf("video:metadata:%s", filePath)
}

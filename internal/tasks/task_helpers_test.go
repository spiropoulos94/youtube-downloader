package tasks

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hibiken/asynq"
)

func TestLogTaskInfo(t *testing.T) {
	// Redirect log output to a buffer so we can check it
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr) // Reset log output when test ends

	// Create a test payload
	payload := map[string]interface{}{
		"url":      "https://example.com/video.mp4",
		"format":   "mp4",
		"quality":  "720p",
		"filename": "test-video.mp4",
	}
	payloadBytes, _ := json.Marshal(payload)

	// Create a test result
	result := map[string]interface{}{
		"status":   "completed",
		"duration": 125.5,
		"size":     "15.2MB",
	}
	resultBytes, _ := json.Marshal(result)

	// Create a mock task info
	now := time.Now()
	taskInfo := &asynq.TaskInfo{
		ID:            "task123",
		Queue:         "default",
		Type:          "video:download",
		Payload:       payloadBytes,
		State:         asynq.TaskStateCompleted,
		MaxRetry:      3,
		Retried:       1,
		LastErr:       "timeout error",
		LastFailedAt:  now.Add(-1 * time.Hour),
		Timeout:       5 * time.Minute,
		Deadline:      now.Add(1 * time.Hour),
		NextProcessAt: now.Add(5 * time.Minute),
		Result:        resultBytes,
	}

	// Call the function
	LogTaskInfo(taskInfo)

	// Verify the log output contains expected information
	logOutput := buf.String()

	// Check for key fields in the output
	expectedFields := []string{
		"Decoded Payload", "url", "format", "quality", "filename",
		"Decoded Result", "status", "completed", "duration", "size",
		"ID", "task123",
		"Queue", "default",
		"Type", "video:download",
		"State",
		"MaxRetry", "3",
		"Retried", "1",
		"LastErr", "timeout error",
	}

	for _, field := range expectedFields {
		if !strings.Contains(logOutput, field) {
			t.Errorf("Expected log output to contain %q, but it didn't.\nLog output: %s", field, logOutput)
		}
	}
}

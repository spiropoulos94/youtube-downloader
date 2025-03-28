package tasks

import (
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"
)

// LogTaskInfo logs detailed information about a task including its payload and result
func LogTaskInfo(info *asynq.TaskInfo) {
	var payloadData, resultData map[string]interface{}
	json.Unmarshal(info.Payload, &payloadData)
	json.Unmarshal(info.Result, &resultData)

	// Log decoded values
	log.Printf("Decoded Payload: %+v", payloadData)
	log.Printf("Decoded Result: %+v", resultData)

	infoJSON, _ := json.Marshal(map[string]interface{}{
		"ID":            info.ID,
		"Queue":         info.Queue,
		"Type":          info.Type,
		"Payload":       payloadData,
		"State":         info.State,
		"MaxRetry":      info.MaxRetry,
		"Retried":       info.Retried,
		"LastErr":       info.LastErr,
		"LastFailedAt":  info.LastFailedAt,
		"Timeout":       info.Timeout,
		"Deadline":      info.Deadline,
		"Group":         info.Group,
		"NextProcessAt": info.NextProcessAt,
		"Result":        resultData,
	})
	log.Printf("Task info: %s", string(infoJSON))
}

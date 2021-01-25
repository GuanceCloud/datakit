/*
 * Kafka
 *
 * Kafka Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ShowBackgroundTaskRequest struct {
	ProjectId  string `json:"project_id"`
	InstanceId string `json:"instance_id"`
	TaskId     string `json:"task_id"`
}

func (o ShowBackgroundTaskRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowBackgroundTaskRequest struct{}"
	}

	return strings.Join([]string{"ShowBackgroundTaskRequest", string(data)}, " ")
}

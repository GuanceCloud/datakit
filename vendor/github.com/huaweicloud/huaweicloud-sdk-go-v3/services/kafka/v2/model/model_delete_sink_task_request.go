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
type DeleteSinkTaskRequest struct {
	ProjectId   string `json:"project_id"`
	ConnectorId string `json:"connector_id"`
	TaskId      string `json:"task_id"`
}

func (o DeleteSinkTaskRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteSinkTaskRequest struct{}"
	}

	return strings.Join([]string{"DeleteSinkTaskRequest", string(data)}, " ")
}

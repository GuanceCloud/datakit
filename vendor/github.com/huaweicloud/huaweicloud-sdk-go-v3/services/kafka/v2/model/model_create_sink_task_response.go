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

// Response Object
type CreateSinkTaskResponse struct {
	// 任务ID。
	TaskId         *string `json:"task_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o CreateSinkTaskResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateSinkTaskResponse struct{}"
	}

	return strings.Join([]string{"CreateSinkTaskResponse", string(data)}, " ")
}

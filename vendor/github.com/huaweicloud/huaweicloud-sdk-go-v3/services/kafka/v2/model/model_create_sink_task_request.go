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
type CreateSinkTaskRequest struct {
	ProjectId   string             `json:"project_id"`
	ConnectorId string             `json:"connector_id"`
	Body        *CreateSinkTaskReq `json:"body,omitempty"`
}

func (o CreateSinkTaskRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateSinkTaskRequest struct{}"
	}

	return strings.Join([]string{"CreateSinkTaskRequest", string(data)}, " ")
}

/*
 * RabbitMQ
 *
 * RabbitMQ Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type BatchRestartOrDeleteInstancesRequest struct {
	ProjectId string                           `json:"project_id"`
	Body      *BatchRestartOrDeleteInstanceReq `json:"body,omitempty"`
}

func (o BatchRestartOrDeleteInstancesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchRestartOrDeleteInstancesRequest struct{}"
	}

	return strings.Join([]string{"BatchRestartOrDeleteInstancesRequest", string(data)}, " ")
}

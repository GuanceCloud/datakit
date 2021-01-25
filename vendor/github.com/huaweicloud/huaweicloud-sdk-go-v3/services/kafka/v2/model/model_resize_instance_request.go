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
type ResizeInstanceRequest struct {
	ProjectId  string             `json:"project_id"`
	InstanceId string             `json:"instance_id"`
	Body       *ResizeInstanceReq `json:"body,omitempty"`
}

func (o ResizeInstanceRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResizeInstanceRequest struct{}"
	}

	return strings.Join([]string{"ResizeInstanceRequest", string(data)}, " ")
}

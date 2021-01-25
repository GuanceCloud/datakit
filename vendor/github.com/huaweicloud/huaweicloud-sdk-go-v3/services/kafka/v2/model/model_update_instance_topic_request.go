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
type UpdateInstanceTopicRequest struct {
	ProjectId  string                  `json:"project_id"`
	InstanceId string                  `json:"instance_id"`
	Body       *UpdateInstanceTopicReq `json:"body,omitempty"`
}

func (o UpdateInstanceTopicRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateInstanceTopicRequest struct{}"
	}

	return strings.Join([]string{"UpdateInstanceTopicRequest", string(data)}, " ")
}

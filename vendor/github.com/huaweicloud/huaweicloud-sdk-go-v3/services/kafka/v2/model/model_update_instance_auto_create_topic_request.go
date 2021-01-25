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
type UpdateInstanceAutoCreateTopicRequest struct {
	ProjectId  string                            `json:"project_id"`
	InstanceId string                            `json:"instance_id"`
	Body       *UpdateInstanceAutoCreateTopicReq `json:"body,omitempty"`
}

func (o UpdateInstanceAutoCreateTopicRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateInstanceAutoCreateTopicRequest struct{}"
	}

	return strings.Join([]string{"UpdateInstanceAutoCreateTopicRequest", string(data)}, " ")
}

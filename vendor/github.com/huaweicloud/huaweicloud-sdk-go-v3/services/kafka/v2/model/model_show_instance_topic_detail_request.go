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
type ShowInstanceTopicDetailRequest struct {
	ProjectId  string `json:"project_id"`
	InstanceId string `json:"instance_id"`
	Topic      string `json:"topic"`
}

func (o ShowInstanceTopicDetailRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowInstanceTopicDetailRequest struct{}"
	}

	return strings.Join([]string{"ShowInstanceTopicDetailRequest", string(data)}, " ")
}

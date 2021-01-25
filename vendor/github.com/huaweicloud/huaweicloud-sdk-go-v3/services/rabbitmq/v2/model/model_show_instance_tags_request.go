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
type ShowInstanceTagsRequest struct {
	ProjectId  string `json:"project_id"`
	InstanceId string `json:"instance_id"`
}

func (o ShowInstanceTagsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowInstanceTagsRequest struct{}"
	}

	return strings.Join([]string{"ShowInstanceTagsRequest", string(data)}, " ")
}

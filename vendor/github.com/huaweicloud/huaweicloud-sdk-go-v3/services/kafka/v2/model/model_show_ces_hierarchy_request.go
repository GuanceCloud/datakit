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
type ShowCesHierarchyRequest struct {
	ProjectId  string `json:"project_id"`
	InstanceId string `json:"instance_id"`
}

func (o ShowCesHierarchyRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowCesHierarchyRequest struct{}"
	}

	return strings.Join([]string{"ShowCesHierarchyRequest", string(data)}, " ")
}

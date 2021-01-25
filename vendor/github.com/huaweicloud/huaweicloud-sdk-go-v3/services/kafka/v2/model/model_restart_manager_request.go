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
type RestartManagerRequest struct {
	ProjectId  string `json:"project_id"`
	InstanceId string `json:"instance_id"`
}

func (o RestartManagerRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RestartManagerRequest struct{}"
	}

	return strings.Join([]string{"RestartManagerRequest", string(data)}, " ")
}

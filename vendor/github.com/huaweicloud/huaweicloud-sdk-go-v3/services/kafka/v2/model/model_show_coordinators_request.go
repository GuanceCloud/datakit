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
type ShowCoordinatorsRequest struct {
	ProjectId  string `json:"project_id"`
	InstanceId string `json:"instance_id"`
}

func (o ShowCoordinatorsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowCoordinatorsRequest struct{}"
	}

	return strings.Join([]string{"ShowCoordinatorsRequest", string(data)}, " ")
}

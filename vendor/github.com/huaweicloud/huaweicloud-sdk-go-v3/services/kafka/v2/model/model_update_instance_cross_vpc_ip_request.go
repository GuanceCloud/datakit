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
type UpdateInstanceCrossVpcIpRequest struct {
	ProjectId  string                       `json:"project_id"`
	InstanceId string                       `json:"instance_id"`
	Body       *UpdateInstanceCrossVpcIpReq `json:"body,omitempty"`
}

func (o UpdateInstanceCrossVpcIpRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateInstanceCrossVpcIpRequest struct{}"
	}

	return strings.Join([]string{"UpdateInstanceCrossVpcIpRequest", string(data)}, " ")
}

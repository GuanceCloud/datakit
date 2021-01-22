/*
 * ServiceStage
 *
 * ServiceStage的API,包括应用管理和仓库授权管理
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type UpdateInstanceActionRequest struct {
	ApplicationId string          `json:"application_id"`
	ComponentId   string          `json:"component_id"`
	InstanceId    string          `json:"instance_id"`
	Body          *InstanceAction `json:"body,omitempty"`
}

func (o UpdateInstanceActionRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateInstanceActionRequest struct{}"
	}

	return strings.Join([]string{"UpdateInstanceActionRequest", string(data)}, " ")
}

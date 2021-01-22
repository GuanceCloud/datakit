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
type ChangeInstanceRequest struct {
	ApplicationId string          `json:"application_id"`
	ComponentId   string          `json:"component_id"`
	InstanceId    string          `json:"instance_id"`
	Body          *InstanceModify `json:"body,omitempty"`
}

func (o ChangeInstanceRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ChangeInstanceRequest struct{}"
	}

	return strings.Join([]string{"ChangeInstanceRequest", string(data)}, " ")
}

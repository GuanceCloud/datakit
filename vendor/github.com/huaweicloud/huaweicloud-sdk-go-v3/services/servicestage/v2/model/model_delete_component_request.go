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
type DeleteComponentRequest struct {
	ApplicationId string `json:"application_id"`
	ComponentId   string `json:"component_id"`
	Force         *bool  `json:"force,omitempty"`
}

func (o DeleteComponentRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteComponentRequest struct{}"
	}

	return strings.Join([]string{"DeleteComponentRequest", string(data)}, " ")
}

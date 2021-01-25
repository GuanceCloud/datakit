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
type ChangeComponentRequest struct {
	ApplicationId string           `json:"application_id"`
	ComponentId   string           `json:"component_id"`
	Body          *ComponentModify `json:"body,omitempty"`
}

func (o ChangeComponentRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ChangeComponentRequest struct{}"
	}

	return strings.Join([]string{"ChangeComponentRequest", string(data)}, " ")
}

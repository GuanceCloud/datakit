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
type CreateComponentRequest struct {
	ApplicationId string           `json:"application_id"`
	Body          *ComponentCreate `json:"body,omitempty"`
}

func (o CreateComponentRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateComponentRequest struct{}"
	}

	return strings.Join([]string{"CreateComponentRequest", string(data)}, " ")
}

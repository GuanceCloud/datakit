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
type ChangeResourceInEnvironmentRequest struct {
	EnvironmentId string                     `json:"environment_id"`
	Body          *EnvironmentResourceModify `json:"body,omitempty"`
}

func (o ChangeResourceInEnvironmentRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ChangeResourceInEnvironmentRequest struct{}"
	}

	return strings.Join([]string{"ChangeResourceInEnvironmentRequest", string(data)}, " ")
}

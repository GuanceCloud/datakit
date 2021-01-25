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

// Response Object
type CreatePasswordAuthResponse struct {
	Authorization  *AuthorizationVo `json:"authorization,omitempty"`
	HttpStatusCode int              `json:"-"`
}

func (o CreatePasswordAuthResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreatePasswordAuthResponse struct{}"
	}

	return strings.Join([]string{"CreatePasswordAuthResponse", string(data)}, " ")
}

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
type CreatePersonalAuthResponse struct {
	Authorization  *AuthorizationVo `json:"authorization,omitempty"`
	HttpStatusCode int              `json:"-"`
}

func (o CreatePersonalAuthResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreatePersonalAuthResponse struct{}"
	}

	return strings.Join([]string{"CreatePersonalAuthResponse", string(data)}, " ")
}

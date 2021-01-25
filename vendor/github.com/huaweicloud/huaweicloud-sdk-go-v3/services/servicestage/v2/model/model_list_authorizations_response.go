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
type ListAuthorizationsResponse struct {
	// 授权列表。
	Authorizations *[]AuthorizationVo `json:"authorizations,omitempty"`
	HttpStatusCode int                `json:"-"`
}

func (o ListAuthorizationsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListAuthorizationsResponse struct{}"
	}

	return strings.Join([]string{"ListAuthorizationsResponse", string(data)}, " ")
}

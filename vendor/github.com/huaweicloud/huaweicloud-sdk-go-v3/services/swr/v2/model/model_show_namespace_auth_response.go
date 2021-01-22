/*
 * SWR
 *
 * SWR API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ShowNamespaceAuthResponse struct {
	// id
	Id *int32 `json:"id,omitempty"`
	// 组织名称
	Name *string `json:"name,omitempty"`
	// 组织创建者
	CreatorName *string   `json:"creator_name,omitempty"`
	SelfAuth    *UserAuth `json:"self_auth,omitempty"`
	// 其他用户的权限
	OthersAuths    *[]UserAuth `json:"others_auths,omitempty"`
	HttpStatusCode int         `json:"-"`
}

func (o ShowNamespaceAuthResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowNamespaceAuthResponse struct{}"
	}

	return strings.Join([]string{"ShowNamespaceAuthResponse", string(data)}, " ")
}

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
type ShowNamespaceResponse struct {
	// id
	Id *int32 `json:"id,omitempty"`
	// 组织名称
	Name *string `json:"name,omitempty"`
	// IAM用户名
	CreatorName *string `json:"creator_name,omitempty"`
	// 用户权限。7表示管理权限，3表示编辑权限，1表示读取权限。
	Auth           *int32 `json:"auth,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ShowNamespaceResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowNamespaceResponse struct{}"
	}

	return strings.Join([]string{"ShowNamespaceResponse", string(data)}, " ")
}

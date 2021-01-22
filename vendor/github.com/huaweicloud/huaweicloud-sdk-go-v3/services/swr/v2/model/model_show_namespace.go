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

type ShowNamespace struct {
	// id
	Id int32 `json:"id"`
	// 组织名称
	Name string `json:"name"`
	// IAM用户名
	CreatorName string `json:"creator_name"`
	// 用户权限。7表示管理权限，3表示编辑权限，1表示读取权限。
	Auth int32 `json:"auth"`
}

func (o ShowNamespace) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowNamespace struct{}"
	}

	return strings.Join([]string{"ShowNamespace", string(data)}, " ")
}

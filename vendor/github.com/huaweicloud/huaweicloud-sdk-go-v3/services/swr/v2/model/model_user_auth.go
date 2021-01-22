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

type UserAuth struct {
	// 用户id，需要从IAM服务获取
	UserId string `json:"user_id"`
	// 用户名，需要从IAM服务获取
	UserName string `json:"user_name"`
	// 用户权限，7表示管理权限，3表示编辑权限，1表示读取权限
	Auth int32 `json:"auth"`
}

func (o UserAuth) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UserAuth struct{}"
	}

	return strings.Join([]string{"UserAuth", string(data)}, " ")
}

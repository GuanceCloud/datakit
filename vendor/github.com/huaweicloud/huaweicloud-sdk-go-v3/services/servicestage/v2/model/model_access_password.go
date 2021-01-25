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

type AccessPassword struct {
	// 授权名称。
	Name string `json:"name"`
	// 仓库用户名。
	User string `json:"user"`
	// 仓库密码。
	Password string `json:"password"`
}

func (o AccessPassword) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AccessPassword struct{}"
	}

	return strings.Join([]string{"AccessPassword", string(data)}, " ")
}

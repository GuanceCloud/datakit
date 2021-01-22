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

type AccessToken struct {
	// 授权名称。
	Name string `json:"name"`
	// git仓库设置中创建的私有token。
	Token string `json:"token"`
	// git仓库的主机地址，如https://192.168.1.1:8080/gitlab，默认为官方主机。
	Host *string `json:"host,omitempty"`
}

func (o AccessToken) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AccessToken struct{}"
	}

	return strings.Join([]string{"AccessToken", string(data)}, " ")
}

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

type OAuth struct {
	// 授权名称。
	Name string `json:"name"`
	// git仓库授权后，重定向回来的url里面的query参数。
	Code string `json:"code"`
	// git仓库授权后，一次性的认证编码和随机串。
	State string `json:"state"`
}

func (o OAuth) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "OAuth struct{}"
	}

	return strings.Join([]string{"OAuth", string(data)}, " ")
}

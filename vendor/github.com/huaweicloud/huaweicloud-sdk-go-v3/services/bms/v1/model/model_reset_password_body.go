/*
 * BMS
 *
 * BMS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 一键重置裸金属服务器密码接口请求结构体
type ResetPasswordBody struct {
	ResetPassword *ResetPassword `json:"reset-password"`
}

func (o ResetPasswordBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResetPasswordBody struct{}"
	}

	return strings.Join([]string{"ResetPasswordBody", string(data)}, " ")
}

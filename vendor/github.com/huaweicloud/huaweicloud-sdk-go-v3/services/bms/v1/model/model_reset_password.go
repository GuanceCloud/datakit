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

// reset-password信息详情
type ResetPassword struct {
	// 裸金属服务器新密码。该接口不做密码安全性校验，设置的密码复杂度请遵循密码规则。密码规则：密码长度范围为8到26位。密码至少包含以下4种字符中的3种：大写字母小写字母数字特殊字符Windows：!@$%-_=+[]:./?Linux：!@%^-_=+[]{}:,./?密码不能包含用户名或用户名的逆序。Windows系统的裸金属服务器，不能包含用户名中超过两个连续字符的部分。
	NewPassword string `json:"new_password"`
}

func (o ResetPassword) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResetPassword struct{}"
	}

	return strings.Join([]string{"ResetPassword", string(data)}, " ")
}

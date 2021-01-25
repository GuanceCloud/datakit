/*
 * RabbitMQ
 *
 * RabbitMQ Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type ResetPasswordReq struct {
	// 8-32个字符。 至少包含以下字符中的3种： - 大写字母 - 小写字母 - 数字 - 特殊字符`~!@#$%^&*()-_=+\\\\|[{}];:\\'\\\",<.>/?  和空格，并且不能以-开头。
	NewPassword *string `json:"new_password,omitempty"`
}

func (o ResetPasswordReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResetPasswordReq struct{}"
	}

	return strings.Join([]string{"ResetPasswordReq", string(data)}, " ")
}

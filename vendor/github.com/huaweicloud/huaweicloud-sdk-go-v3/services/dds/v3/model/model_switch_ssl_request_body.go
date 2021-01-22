/*
 * DDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type SwitchSslRequestBody struct {
	// SSL开关选项。取值：取“0”，表示DDS实例默认不启用SSL连接。取“1”，表示DDS实例默认启用SSL连接。
	SslOption string `json:"ssl_option"`
}

func (o SwitchSslRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SwitchSslRequestBody struct{}"
	}

	return strings.Join([]string{"SwitchSslRequestBody", string(data)}, " ")
}

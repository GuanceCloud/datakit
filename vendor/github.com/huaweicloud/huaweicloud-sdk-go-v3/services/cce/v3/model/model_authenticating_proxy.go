/*
 * CCE
 *
 * CCE开放API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// authenticatingProxy模式相关配置。认证模式为authenticating_proxy时必选
type AuthenticatingProxy struct {
	// authenticating_proxy模式配置的x509格式CA证书(base64编码)。 最大长度：1M
	Ca *string `json:"ca,omitempty"`
}

func (o AuthenticatingProxy) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AuthenticatingProxy struct{}"
	}

	return strings.Join([]string{"AuthenticatingProxy", string(data)}, " ")
}

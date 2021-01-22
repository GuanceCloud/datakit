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

type AuthInfo struct {
	// Base64加密的认证信息
	Auth string `json:"auth"`
}

func (o AuthInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AuthInfo struct{}"
	}

	return strings.Join([]string{"AuthInfo", string(data)}, " ")
}

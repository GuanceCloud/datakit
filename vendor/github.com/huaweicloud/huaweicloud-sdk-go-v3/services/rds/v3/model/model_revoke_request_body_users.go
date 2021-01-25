/*
 * RDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type RevokeRequestBodyUsers struct {
	// 数据库用户名称。
	Name string `json:"name"`
}

func (o RevokeRequestBodyUsers) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RevokeRequestBodyUsers struct{}"
	}

	return strings.Join([]string{"RevokeRequestBodyUsers", string(data)}, " ")
}

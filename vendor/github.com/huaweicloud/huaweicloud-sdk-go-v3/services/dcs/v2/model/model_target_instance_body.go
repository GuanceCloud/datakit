/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 目标实例信息。
type TargetInstanceBody struct {
	// Redis实例ID（target_instance信息中必须填写）。
	Id string `json:"id"`
	// Redis实例名称(target_instance信息中填写)。
	Name *string `json:"name,omitempty"`
	// Redis密码，如果设置了密码，则必须填写。
	Password *string `json:"password,omitempty"`
}

func (o TargetInstanceBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "TargetInstanceBody struct{}"
	}

	return strings.Join([]string{"TargetInstanceBody", string(data)}, " ")
}

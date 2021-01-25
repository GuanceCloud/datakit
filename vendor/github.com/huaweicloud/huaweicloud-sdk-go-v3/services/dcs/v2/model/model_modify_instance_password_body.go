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

type ModifyInstancePasswordBody struct {
	// 旧密码
	OldPassword *string `json:"old_password,omitempty"`
	// 新密码
	NewPassword *string `json:"new_password,omitempty"`
}

func (o ModifyInstancePasswordBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ModifyInstancePasswordBody struct{}"
	}

	return strings.Join([]string{"ModifyInstancePasswordBody", string(data)}, " ")
}

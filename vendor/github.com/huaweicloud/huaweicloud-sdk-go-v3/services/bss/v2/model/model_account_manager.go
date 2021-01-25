/*
 * BSS
 *
 * Business Support System API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type AccountManager struct {
	// |参数名称：客户经理登录名称。| |参数约束及描述：客户经理登录名称。最大长度128，必填|
	AccountName *string `json:"account_name,omitempty"`
}

func (o AccountManager) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AccountManager struct{}"
	}

	return strings.Join([]string{"AccountManager", string(data)}, " ")
}

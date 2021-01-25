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

type CheckSubcustomerUserReq struct {
	// |参数名称：该字段内容可填为：“email”、“mobile”或“name”| |参数的约束及描述：该参数必填，且只允许字符串|
	SearchType string `json:"search_type"`
	// |参数名称：手机、邮箱或用户名| |参数的约束及描述：该参数必填，且只允许字符串,手机包括国家码，以00开头，格式：00XX-XXXXXXXX。目前手机号仅仅支持以86为国家码|
	SearchValue string `json:"search_value"`
}

func (o CheckSubcustomerUserReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CheckSubcustomerUserReq struct{}"
	}

	return strings.Join([]string{"CheckSubcustomerUserReq", string(data)}, " ")
}

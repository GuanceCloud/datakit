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

type EnterprisePersonNew struct {
	// |参数名称：法人姓名| |参数的约束及描述：该参数必填，且只允许字符串|
	LegelName string `json:"legel_name"`
	// |参数名称：法人身份证号| |参数的约束及描述：该参数必填，且只允许字符串|
	LegelIdNumber string `json:"legel_id_number"`
	// |参数名称：认证人角色| |参数的约束及描述：该参数非必填，legalPerson ：法人代表 authorizedPerson：授权人|
	CertifierRole *string `json:"certifier_role,omitempty"`
}

func (o EnterprisePersonNew) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "EnterprisePersonNew struct{}"
	}

	return strings.Join([]string{"EnterprisePersonNew", string(data)}, " ")
}

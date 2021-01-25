/*
 * VPC
 *
 * VPC Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListSecurityGroupRulesResponse struct {
	// 请求ID
	RequestId *string `json:"request_id,omitempty"`
	// 安全组规则列表响应体
	SecurityGroupRules *[]SecurityGroupRule `json:"security_group_rules,omitempty"`
	PageInfo           *PageInfo            `json:"page_info,omitempty"`
	HttpStatusCode     int                  `json:"-"`
}

func (o ListSecurityGroupRulesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListSecurityGroupRulesResponse struct{}"
	}

	return strings.Join([]string{"ListSecurityGroupRulesResponse", string(data)}, " ")
}

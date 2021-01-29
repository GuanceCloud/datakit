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
type CreateSecurityGroupRuleResponse struct {
	// 请求ID
	RequestId         *string            `json:"request_id,omitempty"`
	SecurityGroupRule *SecurityGroupRule `json:"security_group_rule,omitempty"`
	HttpStatusCode    int                `json:"-"`
}

func (o CreateSecurityGroupRuleResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateSecurityGroupRuleResponse struct{}"
	}

	return strings.Join([]string{"CreateSecurityGroupRuleResponse", string(data)}, " ")
}

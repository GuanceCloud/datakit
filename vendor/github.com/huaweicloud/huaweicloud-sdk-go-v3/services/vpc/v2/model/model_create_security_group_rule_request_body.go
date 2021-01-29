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

//
type CreateSecurityGroupRuleRequestBody struct {
	SecurityGroupRule *CreateSecurityGroupRuleOption `json:"security_group_rule"`
}

func (o CreateSecurityGroupRuleRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateSecurityGroupRuleRequestBody struct{}"
	}

	return strings.Join([]string{"CreateSecurityGroupRuleRequestBody", string(data)}, " ")
}

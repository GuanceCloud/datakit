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

// Request Object
type DeleteSecurityGroupRuleRequest struct {
	SecurityGroupRuleId string `json:"security_group_rule_id"`
}

func (o DeleteSecurityGroupRuleRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteSecurityGroupRuleRequest struct{}"
	}

	return strings.Join([]string{"DeleteSecurityGroupRuleRequest", string(data)}, " ")
}

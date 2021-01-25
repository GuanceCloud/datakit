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
type ShowSecurityGroupRuleRequest struct {
	SecurityGroupRuleId string `json:"security_group_rule_id"`
}

func (o ShowSecurityGroupRuleRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowSecurityGroupRuleRequest struct{}"
	}

	return strings.Join([]string{"ShowSecurityGroupRuleRequest", string(data)}, " ")
}

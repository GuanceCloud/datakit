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
type NeutronRemoveFirewallRuleRequestBody struct {
	// 功能说明：待移除的ACL规则ID
	FirewallRuleId string `json:"firewall_rule_id"`
}

func (o NeutronRemoveFirewallRuleRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronRemoveFirewallRuleRequestBody struct{}"
	}

	return strings.Join([]string{"NeutronRemoveFirewallRuleRequestBody", string(data)}, " ")
}

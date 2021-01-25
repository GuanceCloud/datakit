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
type NeutronShowFirewallRuleRequest struct {
	FirewallRuleId string `json:"firewall_rule_id"`
}

func (o NeutronShowFirewallRuleRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronShowFirewallRuleRequest struct{}"
	}

	return strings.Join([]string{"NeutronShowFirewallRuleRequest", string(data)}, " ")
}

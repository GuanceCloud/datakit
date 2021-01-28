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
type NeutronDeleteFirewallRuleRequest struct {
	FirewallRuleId string `json:"firewall_rule_id"`
}

func (o NeutronDeleteFirewallRuleRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronDeleteFirewallRuleRequest struct{}"
	}

	return strings.Join([]string{"NeutronDeleteFirewallRuleRequest", string(data)}, " ")
}

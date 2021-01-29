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
type NeutronUpdateFirewallRuleRequest struct {
	FirewallRuleId string                                `json:"firewall_rule_id"`
	Body           *NeutronUpdateFirewallRuleRequestBody `json:"body,omitempty"`
}

func (o NeutronUpdateFirewallRuleRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronUpdateFirewallRuleRequest struct{}"
	}

	return strings.Join([]string{"NeutronUpdateFirewallRuleRequest", string(data)}, " ")
}

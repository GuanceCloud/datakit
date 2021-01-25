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
type NeutronShowFirewallRuleResponse struct {
	FirewallRule   *NeutronFirewallRule `json:"firewall_rule,omitempty"`
	HttpStatusCode int                  `json:"-"`
}

func (o NeutronShowFirewallRuleResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronShowFirewallRuleResponse struct{}"
	}

	return strings.Join([]string{"NeutronShowFirewallRuleResponse", string(data)}, " ")
}

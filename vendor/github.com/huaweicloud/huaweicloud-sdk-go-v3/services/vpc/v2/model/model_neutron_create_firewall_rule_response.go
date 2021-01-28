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
type NeutronCreateFirewallRuleResponse struct {
	FirewallRule   *NeutronFirewallRule `json:"firewall_rule,omitempty"`
	HttpStatusCode int                  `json:"-"`
}

func (o NeutronCreateFirewallRuleResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronCreateFirewallRuleResponse struct{}"
	}

	return strings.Join([]string{"NeutronCreateFirewallRuleResponse", string(data)}, " ")
}

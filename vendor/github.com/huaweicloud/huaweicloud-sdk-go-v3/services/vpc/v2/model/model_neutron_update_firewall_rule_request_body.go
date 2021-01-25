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
type NeutronUpdateFirewallRuleRequestBody struct {
	FirewallRule *NeutronUpdateFirewallRuleOption `json:"firewall_rule"`
}

func (o NeutronUpdateFirewallRuleRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronUpdateFirewallRuleRequestBody struct{}"
	}

	return strings.Join([]string{"NeutronUpdateFirewallRuleRequestBody", string(data)}, " ")
}

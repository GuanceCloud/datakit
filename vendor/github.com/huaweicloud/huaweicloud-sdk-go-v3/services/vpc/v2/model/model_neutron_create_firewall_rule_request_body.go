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
type NeutronCreateFirewallRuleRequestBody struct {
	FirewallRule *NeutronCreateFirewallRuleOption `json:"firewall_rule"`
}

func (o NeutronCreateFirewallRuleRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronCreateFirewallRuleRequestBody struct{}"
	}

	return strings.Join([]string{"NeutronCreateFirewallRuleRequestBody", string(data)}, " ")
}

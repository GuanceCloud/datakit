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
type NeutronUpdateFirewallPolicyRequestBody struct {
	FirewallPolicy *NeutronUpdateFirewallPolicyOption `json:"firewall_policy"`
}

func (o NeutronUpdateFirewallPolicyRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronUpdateFirewallPolicyRequestBody struct{}"
	}

	return strings.Join([]string{"NeutronUpdateFirewallPolicyRequestBody", string(data)}, " ")
}

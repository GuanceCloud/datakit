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
type NeutronCreateFirewallPolicyRequestBody struct {
	FirewallPolicy *NeutronCreateFirewallPolicyOption `json:"firewall_policy"`
}

func (o NeutronCreateFirewallPolicyRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronCreateFirewallPolicyRequestBody struct{}"
	}

	return strings.Join([]string{"NeutronCreateFirewallPolicyRequestBody", string(data)}, " ")
}

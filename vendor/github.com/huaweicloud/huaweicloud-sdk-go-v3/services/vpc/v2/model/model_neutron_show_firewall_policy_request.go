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
type NeutronShowFirewallPolicyRequest struct {
	FirewallPolicyId string `json:"firewall_policy_id"`
}

func (o NeutronShowFirewallPolicyRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronShowFirewallPolicyRequest struct{}"
	}

	return strings.Join([]string{"NeutronShowFirewallPolicyRequest", string(data)}, " ")
}

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
type NeutronDeleteFirewallPolicyRequest struct {
	FirewallPolicyId string `json:"firewall_policy_id"`
}

func (o NeutronDeleteFirewallPolicyRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronDeleteFirewallPolicyRequest struct{}"
	}

	return strings.Join([]string{"NeutronDeleteFirewallPolicyRequest", string(data)}, " ")
}

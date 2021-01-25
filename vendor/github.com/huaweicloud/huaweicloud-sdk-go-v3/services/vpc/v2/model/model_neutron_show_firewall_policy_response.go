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
type NeutronShowFirewallPolicyResponse struct {
	FirewallPolicy *NeutronFirewallPolicy `json:"firewall_policy,omitempty"`
	HttpStatusCode int                    `json:"-"`
}

func (o NeutronShowFirewallPolicyResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronShowFirewallPolicyResponse struct{}"
	}

	return strings.Join([]string{"NeutronShowFirewallPolicyResponse", string(data)}, " ")
}

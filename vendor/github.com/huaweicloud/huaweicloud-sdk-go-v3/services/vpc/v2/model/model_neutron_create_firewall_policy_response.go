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
type NeutronCreateFirewallPolicyResponse struct {
	FirewallPolicy *NeutronFirewallPolicy `json:"firewall_policy,omitempty"`
	HttpStatusCode int                    `json:"-"`
}

func (o NeutronCreateFirewallPolicyResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronCreateFirewallPolicyResponse struct{}"
	}

	return strings.Join([]string{"NeutronCreateFirewallPolicyResponse", string(data)}, " ")
}

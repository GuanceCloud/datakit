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
type NeutronUpdateFirewallPolicyRequest struct {
	FirewallPolicyId string                                  `json:"firewall_policy_id"`
	Body             *NeutronUpdateFirewallPolicyRequestBody `json:"body,omitempty"`
}

func (o NeutronUpdateFirewallPolicyRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronUpdateFirewallPolicyRequest struct{}"
	}

	return strings.Join([]string{"NeutronUpdateFirewallPolicyRequest", string(data)}, " ")
}

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
type NeutronAddFirewallRuleRequest struct {
	FirewallPolicyId string                                `json:"firewall_policy_id"`
	Body             *NeutronInsertFirewallRuleRequestBody `json:"body,omitempty"`
}

func (o NeutronAddFirewallRuleRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronAddFirewallRuleRequest struct{}"
	}

	return strings.Join([]string{"NeutronAddFirewallRuleRequest", string(data)}, " ")
}

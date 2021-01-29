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
type NeutronListFirewallPoliciesResponse struct {
	// firewall_policy对象列表
	FirewallPolicies *[]NeutronFirewallPolicy `json:"firewall_policies,omitempty"`
	// 分页信息
	FirewallPoliciesLinks *[]NeutronPageLink `json:"firewall_policies_links,omitempty"`
	HttpStatusCode        int                `json:"-"`
}

func (o NeutronListFirewallPoliciesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronListFirewallPoliciesResponse struct{}"
	}

	return strings.Join([]string{"NeutronListFirewallPoliciesResponse", string(data)}, " ")
}

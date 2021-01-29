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
type NeutronListFirewallGroupsRequest struct {
	Marker                  *string   `json:"marker,omitempty"`
	Limit                   *int32    `json:"limit,omitempty"`
	Id                      *[]string `json:"id,omitempty"`
	Name                    *[]string `json:"name,omitempty"`
	Description             *[]string `json:"description,omitempty"`
	IngressFirewallPolicyId *string   `json:"ingress_firewall_policy_id,omitempty"`
	EgressFirewallPolicyId  *string   `json:"egress_firewall_policy_id,omitempty"`
}

func (o NeutronListFirewallGroupsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronListFirewallGroupsRequest struct{}"
	}

	return strings.Join([]string{"NeutronListFirewallGroupsRequest", string(data)}, " ")
}

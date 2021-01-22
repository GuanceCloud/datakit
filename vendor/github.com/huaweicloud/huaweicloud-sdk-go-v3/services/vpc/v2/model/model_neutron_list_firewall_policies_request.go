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
type NeutronListFirewallPoliciesRequest struct {
	Limit       *int32    `json:"limit,omitempty"`
	Marker      *string   `json:"marker,omitempty"`
	Id          *[]string `json:"id,omitempty"`
	Name        *[]string `json:"name,omitempty"`
	Description *[]string `json:"description,omitempty"`
	TenantId    *string   `json:"tenant_id,omitempty"`
}

func (o NeutronListFirewallPoliciesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronListFirewallPoliciesRequest struct{}"
	}

	return strings.Join([]string{"NeutronListFirewallPoliciesRequest", string(data)}, " ")
}

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
type NeutronListFirewallRulesRequest struct {
	Marker      *string   `json:"marker,omitempty"`
	Limit       *int32    `json:"limit,omitempty"`
	Id          *[]string `json:"id,omitempty"`
	Name        *[]string `json:"name,omitempty"`
	Description *[]string `json:"description,omitempty"`
	Action      *string   `json:"action,omitempty"`
	TenantId    *string   `json:"tenant_id,omitempty"`
}

func (o NeutronListFirewallRulesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronListFirewallRulesRequest struct{}"
	}

	return strings.Join([]string{"NeutronListFirewallRulesRequest", string(data)}, " ")
}

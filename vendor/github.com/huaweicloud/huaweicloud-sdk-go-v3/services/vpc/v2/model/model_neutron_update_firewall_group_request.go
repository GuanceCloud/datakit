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
type NeutronUpdateFirewallGroupRequest struct {
	FirewallGroupId string                                 `json:"firewall_group_id"`
	Body            *NeutronUpdateFirewallGroupRequestBody `json:"body,omitempty"`
}

func (o NeutronUpdateFirewallGroupRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronUpdateFirewallGroupRequest struct{}"
	}

	return strings.Join([]string{"NeutronUpdateFirewallGroupRequest", string(data)}, " ")
}

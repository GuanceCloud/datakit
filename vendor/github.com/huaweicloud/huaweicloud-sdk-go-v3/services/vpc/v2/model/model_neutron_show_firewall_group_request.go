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
type NeutronShowFirewallGroupRequest struct {
	FirewallGroupId string `json:"firewall_group_id"`
}

func (o NeutronShowFirewallGroupRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronShowFirewallGroupRequest struct{}"
	}

	return strings.Join([]string{"NeutronShowFirewallGroupRequest", string(data)}, " ")
}

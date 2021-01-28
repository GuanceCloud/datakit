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
type NeutronDeleteFirewallGroupRequest struct {
	FirewallGroupId string `json:"firewall_group_id"`
}

func (o NeutronDeleteFirewallGroupRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronDeleteFirewallGroupRequest struct{}"
	}

	return strings.Join([]string{"NeutronDeleteFirewallGroupRequest", string(data)}, " ")
}

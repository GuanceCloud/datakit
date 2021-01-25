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
type NeutronCreateFirewallGroupResponse struct {
	FirewallGroup  *NeutronFirewallGroup `json:"firewall_group,omitempty"`
	HttpStatusCode int                   `json:"-"`
}

func (o NeutronCreateFirewallGroupResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronCreateFirewallGroupResponse struct{}"
	}

	return strings.Join([]string{"NeutronCreateFirewallGroupResponse", string(data)}, " ")
}

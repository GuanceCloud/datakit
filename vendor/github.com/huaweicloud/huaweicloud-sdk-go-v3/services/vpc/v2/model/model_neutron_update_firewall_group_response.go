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
type NeutronUpdateFirewallGroupResponse struct {
	FirewallGroup  *NeutronFirewallGroup `json:"firewall_group,omitempty"`
	HttpStatusCode int                   `json:"-"`
}

func (o NeutronUpdateFirewallGroupResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronUpdateFirewallGroupResponse struct{}"
	}

	return strings.Join([]string{"NeutronUpdateFirewallGroupResponse", string(data)}, " ")
}

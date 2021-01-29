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

//
type NeutronUpdateFirewallGroupRequestBody struct {
	FirewallGroup *NeutronUpdateFirewallGroupOption `json:"firewall_group"`
}

func (o NeutronUpdateFirewallGroupRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronUpdateFirewallGroupRequestBody struct{}"
	}

	return strings.Join([]string{"NeutronUpdateFirewallGroupRequestBody", string(data)}, " ")
}

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
type NeutronCreateFirewallGroupRequest struct {
	Body *NeutronCreateFirewallGroupRequestBody `json:"body,omitempty"`
}

func (o NeutronCreateFirewallGroupRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronCreateFirewallGroupRequest struct{}"
	}

	return strings.Join([]string{"NeutronCreateFirewallGroupRequest", string(data)}, " ")
}

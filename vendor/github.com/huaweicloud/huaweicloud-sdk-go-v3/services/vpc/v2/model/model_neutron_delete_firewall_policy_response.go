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
type NeutronDeleteFirewallPolicyResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o NeutronDeleteFirewallPolicyResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronDeleteFirewallPolicyResponse struct{}"
	}

	return strings.Join([]string{"NeutronDeleteFirewallPolicyResponse", string(data)}, " ")
}

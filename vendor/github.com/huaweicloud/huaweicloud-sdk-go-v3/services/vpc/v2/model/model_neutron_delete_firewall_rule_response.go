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
type NeutronDeleteFirewallRuleResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o NeutronDeleteFirewallRuleResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NeutronDeleteFirewallRuleResponse struct{}"
	}

	return strings.Join([]string{"NeutronDeleteFirewallRuleResponse", string(data)}, " ")
}

/*
 * NAT
 *
 * Open Api of Public Nat.
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ShowNatGatewayDnatRuleResponse struct {
	DnatRule       *NatGatewayDnatRuleResponseBody `json:"dnat_rule,omitempty"`
	HttpStatusCode int                             `json:"-"`
}

func (o ShowNatGatewayDnatRuleResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowNatGatewayDnatRuleResponse struct{}"
	}

	return strings.Join([]string{"ShowNatGatewayDnatRuleResponse", string(data)}, " ")
}

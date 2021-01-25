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

// Request Object
type DeleteNatGatewayDnatRuleRequest struct {
	NatGatewayId string `json:"nat_gateway_id"`
	DnatRuleId   string `json:"dnat_rule_id"`
}

func (o DeleteNatGatewayDnatRuleRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteNatGatewayDnatRuleRequest struct{}"
	}

	return strings.Join([]string{"DeleteNatGatewayDnatRuleRequest", string(data)}, " ")
}

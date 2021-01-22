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
type DeleteNatGatewaySnatRuleRequest struct {
	NatGatewayId string `json:"nat_gateway_id"`
	SnatRuleId   string `json:"snat_rule_id"`
}

func (o DeleteNatGatewaySnatRuleRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteNatGatewaySnatRuleRequest struct{}"
	}

	return strings.Join([]string{"DeleteNatGatewaySnatRuleRequest", string(data)}, " ")
}

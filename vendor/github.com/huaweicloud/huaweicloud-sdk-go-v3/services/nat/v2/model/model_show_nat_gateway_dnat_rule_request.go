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
type ShowNatGatewayDnatRuleRequest struct {
	DnatRuleId string `json:"dnat_rule_id"`
}

func (o ShowNatGatewayDnatRuleRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowNatGatewayDnatRuleRequest struct{}"
	}

	return strings.Join([]string{"ShowNatGatewayDnatRuleRequest", string(data)}, " ")
}

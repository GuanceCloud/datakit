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
type ShowNatGatewaySnatRuleResponse struct {
	SnatRule       *NatGatewaySnatRuleResponseBody `json:"snat_rule,omitempty"`
	HttpStatusCode int                             `json:"-"`
}

func (o ShowNatGatewaySnatRuleResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowNatGatewaySnatRuleResponse struct{}"
	}

	return strings.Join([]string{"ShowNatGatewaySnatRuleResponse", string(data)}, " ")
}

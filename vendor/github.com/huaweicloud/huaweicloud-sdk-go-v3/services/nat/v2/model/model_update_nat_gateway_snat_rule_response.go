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
type UpdateNatGatewaySnatRuleResponse struct {
	SnatRule       *NatGatewaySnatRuleResponseBody `json:"snat_rule,omitempty"`
	HttpStatusCode int                             `json:"-"`
}

func (o UpdateNatGatewaySnatRuleResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateNatGatewaySnatRuleResponse struct{}"
	}

	return strings.Join([]string{"UpdateNatGatewaySnatRuleResponse", string(data)}, " ")
}

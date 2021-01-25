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
type CreateNatGatewaySnatRuleResponse struct {
	SnatRule       *NatGatewaySnatRuleResponseBody `json:"snat_rule,omitempty"`
	HttpStatusCode int                             `json:"-"`
}

func (o CreateNatGatewaySnatRuleResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateNatGatewaySnatRuleResponse struct{}"
	}

	return strings.Join([]string{"CreateNatGatewaySnatRuleResponse", string(data)}, " ")
}

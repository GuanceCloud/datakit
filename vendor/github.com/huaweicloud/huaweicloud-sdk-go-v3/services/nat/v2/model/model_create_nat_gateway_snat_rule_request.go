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
type CreateNatGatewaySnatRuleRequest struct {
	Body *CreateNatGatewaySnatRuleRequestOption `json:"body,omitempty"`
}

func (o CreateNatGatewaySnatRuleRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateNatGatewaySnatRuleRequest struct{}"
	}

	return strings.Join([]string{"CreateNatGatewaySnatRuleRequest", string(data)}, " ")
}

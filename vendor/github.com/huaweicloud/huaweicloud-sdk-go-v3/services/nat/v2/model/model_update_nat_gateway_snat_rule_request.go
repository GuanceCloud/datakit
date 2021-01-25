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
type UpdateNatGatewaySnatRuleRequest struct {
	SnatRuleId string                                 `json:"snat_rule_id"`
	Body       *UpdateNatGatewaySnatRuleRequestOption `json:"body,omitempty"`
}

func (o UpdateNatGatewaySnatRuleRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateNatGatewaySnatRuleRequest struct{}"
	}

	return strings.Join([]string{"UpdateNatGatewaySnatRuleRequest", string(data)}, " ")
}

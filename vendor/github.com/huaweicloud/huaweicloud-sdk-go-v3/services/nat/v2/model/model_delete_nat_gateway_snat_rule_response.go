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
type DeleteNatGatewaySnatRuleResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteNatGatewaySnatRuleResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteNatGatewaySnatRuleResponse struct{}"
	}

	return strings.Join([]string{"DeleteNatGatewaySnatRuleResponse", string(data)}, " ")
}

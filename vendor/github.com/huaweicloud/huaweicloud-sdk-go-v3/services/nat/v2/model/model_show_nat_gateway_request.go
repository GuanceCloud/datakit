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
type ShowNatGatewayRequest struct {
	NatGatewayId string `json:"nat_gateway_id"`
}

func (o ShowNatGatewayRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowNatGatewayRequest struct{}"
	}

	return strings.Join([]string{"ShowNatGatewayRequest", string(data)}, " ")
}

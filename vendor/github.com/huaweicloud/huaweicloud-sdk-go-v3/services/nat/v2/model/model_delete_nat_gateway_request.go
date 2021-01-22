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
type DeleteNatGatewayRequest struct {
	NatGatewayId string `json:"nat_gateway_id"`
}

func (o DeleteNatGatewayRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteNatGatewayRequest struct{}"
	}

	return strings.Join([]string{"DeleteNatGatewayRequest", string(data)}, " ")
}

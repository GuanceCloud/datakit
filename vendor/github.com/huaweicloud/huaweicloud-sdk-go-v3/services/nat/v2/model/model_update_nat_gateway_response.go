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
type UpdateNatGatewayResponse struct {
	NatGateway     *NatGatewayResponseBody `json:"nat_gateway,omitempty"`
	HttpStatusCode int                     `json:"-"`
}

func (o UpdateNatGatewayResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateNatGatewayResponse struct{}"
	}

	return strings.Join([]string{"UpdateNatGatewayResponse", string(data)}, " ")
}

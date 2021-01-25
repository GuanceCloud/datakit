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
type CreateNatGatewayResponse struct {
	NatGateway     *NatGatewayResponseBody `json:"nat_gateway,omitempty"`
	HttpStatusCode int                     `json:"-"`
}

func (o CreateNatGatewayResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateNatGatewayResponse struct{}"
	}

	return strings.Join([]string{"CreateNatGatewayResponse", string(data)}, " ")
}

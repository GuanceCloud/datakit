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
type ShowNatGatewayResponse struct {
	NatGateway     *NatGatewayResponseBody `json:"nat_gateway,omitempty"`
	HttpStatusCode int                     `json:"-"`
}

func (o ShowNatGatewayResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowNatGatewayResponse struct{}"
	}

	return strings.Join([]string{"ShowNatGatewayResponse", string(data)}, " ")
}

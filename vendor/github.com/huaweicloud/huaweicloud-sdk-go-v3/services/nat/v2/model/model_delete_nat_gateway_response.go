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
type DeleteNatGatewayResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteNatGatewayResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteNatGatewayResponse struct{}"
	}

	return strings.Join([]string{"DeleteNatGatewayResponse", string(data)}, " ")
}

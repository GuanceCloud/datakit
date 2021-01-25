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

// 创建公网NAT网关实例的请求体。
type CreateNatGatewayRequestBody struct {
	NatGateway *CreateNatGatewayOption `json:"nat_gateway"`
}

func (o CreateNatGatewayRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateNatGatewayRequestBody struct{}"
	}

	return strings.Join([]string{"CreateNatGatewayRequestBody", string(data)}, " ")
}

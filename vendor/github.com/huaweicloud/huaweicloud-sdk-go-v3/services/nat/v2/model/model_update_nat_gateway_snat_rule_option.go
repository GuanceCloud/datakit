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

type UpdateNatGatewaySnatRuleOption struct {
	// 公网NAT网关的id。
	NatGatewayId string `json:"nat_gateway_id"`
	// 功能说明：弹性公网IP，多个弹性公网IP使用逗号分隔。 取值范围：最大长度1024字节。 约束：弹性公网IP的id个数不能超过20个
	PublicIpAddress *string `json:"public_ip_address,omitempty"`
	// SNAT规则的描述，长度限制为255。
	Description *string `json:"description,omitempty"`
}

func (o UpdateNatGatewaySnatRuleOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateNatGatewaySnatRuleOption struct{}"
	}

	return strings.Join([]string{"UpdateNatGatewaySnatRuleOption", string(data)}, " ")
}

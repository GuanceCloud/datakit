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

// 创建SNAT规则的请求体。
type CreateNatGatewaySnatRuleOption struct {
	// 公网NAT网关实例的ID。
	NatGatewayId string `json:"nat_gateway_id"`
	// cidr，可以是网段或者主机格式，与network_id参数二选一。 Source_type=0时，cidr必须是vpc 子网网段的子集(不能相等）; Source_type=1时，cidr必须指定专线侧网段。
	Cidr *string `json:"cidr,omitempty"`
	// 规则使用的网络id。与cidr参数二选一。
	NetworkId *string `json:"network_id,omitempty"`
	// SNAT规则的描述，长度限制为255。
	Description *string `json:"description,omitempty"`
	// 0：VPC侧，可以指定network_id 或者cidr 1：专线侧，只能指定cidr 不输入默认为0（VPC）
	SourceType *int32 `json:"source_type,omitempty"`
	// 功能说明：弹性公网IP的id，多个弹性公网IP使用逗号分隔。 取值范围：最大长度4096字节。 约束：弹性公网IP的id个数不能超过20个。
	FloatingIpId string `json:"floating_ip_id"`
}

func (o CreateNatGatewaySnatRuleOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateNatGatewaySnatRuleOption struct{}"
	}

	return strings.Join([]string{"CreateNatGatewaySnatRuleOption", string(data)}, " ")
}

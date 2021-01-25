/*
 * VPC
 *
 * VPC Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

//
type CreateSecurityGroupRuleOption struct {
	// 功能说明：安全组规则所属的安全组ID
	SecurityGroupId string `json:"security_group_id"`
	// 功能说明：安全组的描述信息 取值范围：0-255个字符，不能包含“<”和“>”
	Description *string `json:"description,omitempty"`
	// 功能说明：安全组规则的出入控制方向 取值范围：ingress 表示入方向；egress 表示出方向
	Direction string `json:"direction"`
	// 功能说明：IP地址协议类型 取值范围：IPv4，IPv6 约束：不填默认值为IPv4
	Ethertype *string `json:"ethertype,omitempty"`
	// 功能说明：协议类型 取值范围：icmp、tcp、udp、icmpv6或IP协议号(0~255) 约束：为空表示支持所有协议。协议为icmpv6时，网络类型应该为IPv6；协议为icmp时，网络类型应该为IPv4
	Protocol *string `json:"protocol,omitempty"`
	// 功能说明：端口取值范围 取值范围：支持单端口(80)，连续端口(1-30)以及不连续端口(22,3389,80) 约束：端口值的范围1~65535
	Multiport *string `json:"multiport,omitempty"`
	// 功能说明：远端IP地址，当direction是egress时为虚拟机访问端的地址，当direction是ingress时为访问虚拟机的地址 取值范围：IP地址，或者cidr格式 约束：与remote_group_id、remote_address_group_id互斥
	RemoteIpPrefix *string `json:"remote_ip_prefix,omitempty"`
	// 功能说明：远端安全组ID，表示该安全组内的流量允许或拒绝 取值范围：租户下存在的安全组ID 约束：与remote_ip_prefix，remote_address_group_id功能互斥
	RemoteGroupId *string `json:"remote_group_id,omitempty"`
	// 功能说明：远端地址组ID 取值范围：租户下存在的地址组ID 约束：与remote_ip_prefix，remote_group_id功能互斥
	RemoteAddressGroupId *string `json:"remote_address_group_id,omitempty"`
	// 功能说明：安全组规则生效策略 取值范围：allow 允许，deny 拒绝 约束：默认值为allow
	Action *string `json:"action,omitempty"`
	// 功能说明：规则在安全组中的优先级 取值范围：1~100，1代表最高优先级 约束：默认值为100
	Priority *string `json:"priority,omitempty"`
}

func (o CreateSecurityGroupRuleOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateSecurityGroupRuleOption struct{}"
	}

	return strings.Join([]string{"CreateSecurityGroupRuleOption", string(data)}, " ")
}

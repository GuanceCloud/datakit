/*
 * VPC
 *
 * VPC Open API
 *
 */

package model

import (
	"encoding/json"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"

	"strings"
)

//
type SecurityGroupRule struct {
	// 功能描述：安全组规则对应的唯一标识 取值范围：带“-”的标准UUID格式
	Id string `json:"id"`
	// 功能说明：安全组的描述信息 取值范围：0-255个字符，不能包含“<”和“>”
	Description string `json:"description"`
	// 功能说明：安全组规则所属的安全组ID
	SecurityGroupId string `json:"security_group_id"`
	// 功能说明：安全组规则的出入控制方向 取值范围：ingress 表示入方向；egress 表示出方向
	Direction string `json:"direction"`
	// 功能说明：协议类型 取值范围：icmp、tcp、udp、icmpv6或IP协议号 约束：为空表示支持所有协议。协议为icmpv6时，网络类型应该为IPv6；协议为icmp时，网络类型应该为IPv4
	Protocol string `json:"protocol"`
	// 功能说明：IP地址协议类型 取值范围：IPv4，IPv6 约束：不填默认值为IPv4
	Ethertype string `json:"ethertype"`
	// 功能说明：端口取值范围 取值范围：支持单端口(80)，连续端口(1-30)以及不连续端口(22,3389,80)
	Multiport string `json:"multiport"`
	// 功能说明：安全组规则生效策略 取值范围：allow 允许，deny 拒绝 约束：默认值为deny
	Action string `json:"action"`
	// 功能说明：优先级 取值范围：1~100，1代表最高优先级
	Priority int32 `json:"priority"`
	// 功能说明：远端安全组ID，表示该安全组内的流量允许或拒绝 取值范围：租户下存在的安全组ID 约束：与remote_ip_prefix，remote_address_group_id功能互斥
	RemoteGroupId string `json:"remote_group_id"`
	// 功能说明：远端IP地址，当direction是egress时为虚拟机访问端的地址，当direction是ingress时为访问虚拟机的地址 取值范围：IP地址，或者cidr格式 约束：与remote_group_id、remote_address_group_id互斥
	RemoteIpPrefix string `json:"remote_ip_prefix"`
	// 功能说明：远端地址组ID 取值范围：租户下存在的地址组ID 约束：与remote_ip_prefix，remote_group_id功能互斥
	RemoteAddressGroupId string `json:"remote_address_group_id"`
	// 功能说明：安全组规则创建时间 取值范围：UTC时间格式：yyyy-MM-ddTHH:mm:ss
	CreatedAt *sdktime.SdkTime `json:"created_at"`
	// 功能说明：安全组规则更新时间 取值范围：UTC时间格式：yyyy-MM-ddTHH:mm:ss
	UpdatedAt *sdktime.SdkTime `json:"updated_at"`
	// 功能说明：安全组规则所属项目ID
	ProjectId string `json:"project_id"`
}

func (o SecurityGroupRule) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SecurityGroupRule struct{}"
	}

	return strings.Join([]string{"SecurityGroupRule", string(data)}, " ")
}

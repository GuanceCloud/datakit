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

// 创建DNAT规则的请求体。
type CreateNatGatewayDnatOption struct {
	// DNAT规则的描述，长度限制为255。
	Description *string `json:"description,omitempty"`
	// 虚拟机或者裸机的Port ID，对应虚拟私有云场景，与private_ip参数二选一。
	PortId *string `json:"port_id,omitempty"`
	// 用户私有IP地址，对应专线、云连接场景，与port_id参数二选一。
	PrivateIp *string `json:"private_ip,omitempty"`
	// 公网NAT网关实例的ID。
	NatGatewayId string `json:"nat_gateway_id"`
	// 虚拟机或者裸机对外提供服务的协议端口号。 取值范围：0~65535。
	InternalServicePort int32 `json:"internal_service_port"`
	// 弹性公网IP的id。
	FloatingIpId string `json:"floating_ip_id"`
	// Floatingip对外提供服务的端口号。 取值范围：0~65535。
	ExternalServicePort int32 `json:"external_service_port"`
	// 协议类型，目前支持TCP/tcp、UDP/udp、ANY/any。 对应协议号6、17、0。
	Protocol string `json:"protocol"`
	// 虚拟机或者裸机对外提供服务的协议端口号范围。 功能说明：该端口范围与external _service_port_range按顺序实现1:1映射。 取值范围：1~65535。 约束：只能以’-’字符连接端口范围。
	InternalServicePortRange *string `json:"internal_service_port_range,omitempty"`
	// Floatingip对外提供服务的端口号范围。 功能说明：该端口范围与internal _service_port_range按顺序实现1:1映射。 取值范围：1~65535。 约束：只能以’-’字符连接端口范围。
	ExternalServicePortRange *string `json:"external_service_port_range,omitempty"`
}

func (o CreateNatGatewayDnatOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateNatGatewayDnatOption struct{}"
	}

	return strings.Join([]string{"CreateNatGatewayDnatOption", string(data)}, " ")
}

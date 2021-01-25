/*
 * NAT
 *
 * Open Api of Public Nat.
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// 更新DNAT规则的请求体。
type UpdateNatGatewayDnatRuleOption struct {
	// NAT网关的id。
	NatGatewayId string `json:"nat_gateway_id"`
	// DNAT规则的描述，长度限制为255。
	Description *string `json:"description,omitempty"`
	// 虚拟机或者裸机的Port ID，对应虚拟私有云场景，与private_ip参数二选一。
	PortId *string `json:"port_id,omitempty"`
	// 用户私有IP地址，对应专线、云连接场景，与port_id参数二选一。
	PrivateIp *string `json:"private_ip,omitempty"`
	// 协议类型，目前支持TCP/tcp、UDP/udp、ANY/any。 对应协议号6、17、0。
	Protocol *UpdateNatGatewayDnatRuleOptionProtocol `json:"protocol,omitempty"`
	// 弹性公网IP的id。
	FloatingIpId *string `json:"floating_ip_id,omitempty"`
	// 虚拟机或者裸机对外提供服务的协议端口号。 取值范围：0~65535。
	InternalServicePort *int32 `json:"internal_service_port,omitempty"`
	// Floatingip对外提供服务的端口号。 取值范围：0~65535。
	ExternalServicePort *int32 `json:"external_service_port,omitempty"`
	// 虚拟机或者裸机对外提供服务的协议端口号范围。 功能说明：该端口范围与external _service_port_range按顺序实现1:1映射。 取值范围：1~65535。 约束：只能以’-’字符连接端口范围。
	InternalServicePortRange *string `json:"internal_service_port_range,omitempty"`
	// Floatingip对外提供服务的端口号范围。 功能说明：该端口范围与internal _service_port_range按顺序实现1:1映射。 取值范围：1~65535。 约束：只能以’-’字符连接端口范围。
	ExternalServicePortRange *string `json:"external_service_port_range,omitempty"`
}

func (o UpdateNatGatewayDnatRuleOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateNatGatewayDnatRuleOption struct{}"
	}

	return strings.Join([]string{"UpdateNatGatewayDnatRuleOption", string(data)}, " ")
}

type UpdateNatGatewayDnatRuleOptionProtocol struct {
	value string
}

type UpdateNatGatewayDnatRuleOptionProtocolEnum struct {
	TCP UpdateNatGatewayDnatRuleOptionProtocol
	UDP UpdateNatGatewayDnatRuleOptionProtocol
	ANY UpdateNatGatewayDnatRuleOptionProtocol
}

func GetUpdateNatGatewayDnatRuleOptionProtocolEnum() UpdateNatGatewayDnatRuleOptionProtocolEnum {
	return UpdateNatGatewayDnatRuleOptionProtocolEnum{
		TCP: UpdateNatGatewayDnatRuleOptionProtocol{
			value: "TCP",
		},
		UDP: UpdateNatGatewayDnatRuleOptionProtocol{
			value: "UDP",
		},
		ANY: UpdateNatGatewayDnatRuleOptionProtocol{
			value: "ANY",
		},
	}
}

func (c UpdateNatGatewayDnatRuleOptionProtocol) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdateNatGatewayDnatRuleOptionProtocol) UnmarshalJSON(b []byte) error {
	myConverter := converter.StringConverterFactory("string")
	if myConverter != nil {
		val, err := myConverter.CovertStringToInterface(strings.Trim(string(b[:]), "\""))
		if err == nil {
			c.value = val.(string)
			return nil
		}
		return err
	} else {
		return errors.New("convert enum data to string error")
	}
}

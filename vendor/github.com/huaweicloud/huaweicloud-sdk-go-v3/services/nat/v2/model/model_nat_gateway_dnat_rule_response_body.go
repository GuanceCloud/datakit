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
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"
	"strings"
)

// DNAT规则的响应体。
type NatGatewayDnatRuleResponseBody struct {
	// DNAT规则的ID。
	Id string `json:"id"`
	// 项目的ID。
	TenantId string `json:"tenant_id"`
	// DNAT规则的描述。长度限制为255。
	Description string `json:"description"`
	// 虚拟机或者裸机的Port ID，对应虚拟私有云场景，与private_ip参数二选一。
	PortId *string `json:"port_id,omitempty"`
	// 用户私有IP地址，对应专线、云连接场景，与port_id参数二选一。
	PrivateIp *string `json:"private_ip,omitempty"`
	// 虚拟机或者裸机对外提供服务的协议端口号。 取值范围：0~65535。
	InternalServicePort int32 `json:"internal_service_port"`
	// 公网NAT网关实例的ID。
	NatGatewayId string `json:"nat_gateway_id"`
	// 弹性公网IP的id。
	FloatingIpId string `json:"floating_ip_id"`
	// 弹性公网IP的IP地址。
	FloatingIpAddress string `json:"floating_ip_address"`
	// Floatingip对外提供服务的端口号。 取值范围：0~65535。
	ExternalServicePort int32 `json:"external_service_port"`
	// 功能说明：DNAT规则的状态。
	Status NatGatewayDnatRuleResponseBodyStatus `json:"status"`
	// 解冻/冻结状态。 取值范围： − “true”： 解冻 − “false”： 冻结
	AdminStateUp bool `json:"admin_state_up"`
	// 虚拟机或者裸机对外提供服务的协议端口号范围。 功能说明：该端口范围与external _service_port_range按顺序实现1:1映射。 取值范围：1~65535。 约束：只能以’-’字符连接端口范围。
	InternalServicePortRange *string `json:"internal_service_port_range,omitempty"`
	// Floatingip对外提供服务的端口号范围。 功能说明：该端口范围与internal _service_port_range按顺序实现1:1映射。 取值范围：1~65535。 约束：只能以’-’字符连接端口范围
	ExternalServicePortRange *string `json:"external_service_port_range,omitempty"`
	// 协议类型，目前支持TCP/tcp、UDP/udp、ANY/any。 对应协议号6、17、0。
	Protocol NatGatewayDnatRuleResponseBodyProtocol `json:"protocol"`
	// DNAT规则的创建时间，遵循UTC时间，格式是yyyy-mm-ddThh:mm:ssZ。
	CreatedAt *sdktime.SdkTime `json:"created_at"`
}

func (o NatGatewayDnatRuleResponseBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NatGatewayDnatRuleResponseBody struct{}"
	}

	return strings.Join([]string{"NatGatewayDnatRuleResponseBody", string(data)}, " ")
}

type NatGatewayDnatRuleResponseBodyStatus struct {
	value string
}

type NatGatewayDnatRuleResponseBodyStatusEnum struct {
	ACTIVE         NatGatewayDnatRuleResponseBodyStatus
	PENDING_CREATE NatGatewayDnatRuleResponseBodyStatus
	PENDING_UPDATE NatGatewayDnatRuleResponseBodyStatus
	PENDING_DELETE NatGatewayDnatRuleResponseBodyStatus
	EIP_FREEZED    NatGatewayDnatRuleResponseBodyStatus
	INACTIVE       NatGatewayDnatRuleResponseBodyStatus
}

func GetNatGatewayDnatRuleResponseBodyStatusEnum() NatGatewayDnatRuleResponseBodyStatusEnum {
	return NatGatewayDnatRuleResponseBodyStatusEnum{
		ACTIVE: NatGatewayDnatRuleResponseBodyStatus{
			value: "ACTIVE",
		},
		PENDING_CREATE: NatGatewayDnatRuleResponseBodyStatus{
			value: "PENDING_CREATE",
		},
		PENDING_UPDATE: NatGatewayDnatRuleResponseBodyStatus{
			value: "PENDING_UPDATE",
		},
		PENDING_DELETE: NatGatewayDnatRuleResponseBodyStatus{
			value: "PENDING_DELETE",
		},
		EIP_FREEZED: NatGatewayDnatRuleResponseBodyStatus{
			value: "EIP_FREEZED",
		},
		INACTIVE: NatGatewayDnatRuleResponseBodyStatus{
			value: "INACTIVE",
		},
	}
}

func (c NatGatewayDnatRuleResponseBodyStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *NatGatewayDnatRuleResponseBodyStatus) UnmarshalJSON(b []byte) error {
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

type NatGatewayDnatRuleResponseBodyProtocol struct {
	value string
}

type NatGatewayDnatRuleResponseBodyProtocolEnum struct {
	TCP NatGatewayDnatRuleResponseBodyProtocol
	UDP NatGatewayDnatRuleResponseBodyProtocol
	ANY NatGatewayDnatRuleResponseBodyProtocol
}

func GetNatGatewayDnatRuleResponseBodyProtocolEnum() NatGatewayDnatRuleResponseBodyProtocolEnum {
	return NatGatewayDnatRuleResponseBodyProtocolEnum{
		TCP: NatGatewayDnatRuleResponseBodyProtocol{
			value: "tcp",
		},
		UDP: NatGatewayDnatRuleResponseBodyProtocol{
			value: "udp",
		},
		ANY: NatGatewayDnatRuleResponseBodyProtocol{
			value: "any",
		},
	}
}

func (c NatGatewayDnatRuleResponseBodyProtocol) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *NatGatewayDnatRuleResponseBodyProtocol) UnmarshalJSON(b []byte) error {
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

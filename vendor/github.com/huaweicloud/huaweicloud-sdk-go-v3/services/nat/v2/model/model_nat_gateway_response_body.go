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

// 公网NAT网关实例的响应体。
type NatGatewayResponseBody struct {
	// 公网NAT网关实例的ID。
	Id string `json:"id"`
	// 项目的ID。
	TenantId string `json:"tenant_id"`
	// 公网NAT网关实例的名字，长度限制为64。
	Name string `json:"name"`
	// 公网NAT网关实例的描述，长度限制为255。
	Description string `json:"description"`
	// 公网NAT网关的规格。 取值为： “1”：小型，SNAT最大连接数10000 “2”：中型，SNAT最大连接数50000 “3”：大型，SNAT最大连接数200000 “4”：超大型，SNAT最大连接数1000000
	Spec NatGatewayResponseBodySpec `json:"spec"`
	// 公网NAT网关实例的状态。
	Status NatGatewayResponseBodyStatus `json:"status"`
	// 解冻/冻结状态。 取值范围： - \"true\"：解冻 - \"false\"：冻结
	AdminStateUp bool `json:"admin_state_up"`
	// 公网NAT网关实例的创建时间，遵循UTC时间，格式是yyyy-mm-ddThh:mm:ssZ。
	CreatedAt *sdktime.SdkTime `json:"created_at"`
	// VPC的id。
	RouterId string `json:"router_id"`
	// 公网NAT网关下行口（DVR的下一跳）所属的network id。
	InternalNetworkId string `json:"internal_network_id"`
	// 企业项目ID。 创建公网NAT网关实例时，关联的企业项目ID。
	EnterpriseProjectId string `json:"enterprise_project_id"`
}

func (o NatGatewayResponseBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NatGatewayResponseBody struct{}"
	}

	return strings.Join([]string{"NatGatewayResponseBody", string(data)}, " ")
}

type NatGatewayResponseBodySpec struct {
	value string
}

type NatGatewayResponseBodySpecEnum struct {
	E_1 NatGatewayResponseBodySpec
	E_2 NatGatewayResponseBodySpec
	E_3 NatGatewayResponseBodySpec
	E_4 NatGatewayResponseBodySpec
}

func GetNatGatewayResponseBodySpecEnum() NatGatewayResponseBodySpecEnum {
	return NatGatewayResponseBodySpecEnum{
		E_1: NatGatewayResponseBodySpec{
			value: "1",
		},
		E_2: NatGatewayResponseBodySpec{
			value: "2",
		},
		E_3: NatGatewayResponseBodySpec{
			value: "3",
		},
		E_4: NatGatewayResponseBodySpec{
			value: "4",
		},
	}
}

func (c NatGatewayResponseBodySpec) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *NatGatewayResponseBodySpec) UnmarshalJSON(b []byte) error {
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

type NatGatewayResponseBodyStatus struct {
	value string
}

type NatGatewayResponseBodyStatusEnum struct {
	ACTIVE         NatGatewayResponseBodyStatus
	PENDING_CREATE NatGatewayResponseBodyStatus
	PENDING_UPDATE NatGatewayResponseBodyStatus
	PENDING_DELETE NatGatewayResponseBodyStatus
	INACTIVE       NatGatewayResponseBodyStatus
}

func GetNatGatewayResponseBodyStatusEnum() NatGatewayResponseBodyStatusEnum {
	return NatGatewayResponseBodyStatusEnum{
		ACTIVE: NatGatewayResponseBodyStatus{
			value: "ACTIVE",
		},
		PENDING_CREATE: NatGatewayResponseBodyStatus{
			value: "PENDING_CREATE",
		},
		PENDING_UPDATE: NatGatewayResponseBodyStatus{
			value: "PENDING_UPDATE",
		},
		PENDING_DELETE: NatGatewayResponseBodyStatus{
			value: "PENDING_DELETE",
		},
		INACTIVE: NatGatewayResponseBodyStatus{
			value: "INACTIVE",
		},
	}
}

func (c NatGatewayResponseBodyStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *NatGatewayResponseBodyStatus) UnmarshalJSON(b []byte) error {
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

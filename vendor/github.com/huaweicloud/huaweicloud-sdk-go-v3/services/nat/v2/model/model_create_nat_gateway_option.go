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

// 创建公网NAT网关实例的请求体。
type CreateNatGatewayOption struct {
	// 公网NAT网关实例的名字，长度限制为64。 公网NAT网关实例的名字仅支持数字、字母、_（下划线）、-（中划线）、中文。
	Name string `json:"name"`
	// VPC的id。
	RouterId string `json:"router_id"`
	// 公网NAT网关下行口（DVR的下一跳）所属的network id。
	InternalNetworkId string `json:"internal_network_id"`
	// 公网NAT网关实例的描述，长度限制为255。
	Description *string `json:"description,omitempty"`
	// 公网NAT网关的规格。 取值为： “1”：小型，SNAT最大连接数10000 “2”：中型，SNAT最大连接数50000 “3”：大型，SNAT最大连接数200000 “4”：超大型，SNAT最大连接数1000000
	Spec CreateNatGatewayOptionSpec `json:"spec"`
	// 企业项目ID 创建公网NAT网关实例时，关联的企业项目ID。 关于企业项目ID的获取及企业项目特性的详细信息，请参考《企业管理用户指南》。
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
}

func (o CreateNatGatewayOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateNatGatewayOption struct{}"
	}

	return strings.Join([]string{"CreateNatGatewayOption", string(data)}, " ")
}

type CreateNatGatewayOptionSpec struct {
	value string
}

type CreateNatGatewayOptionSpecEnum struct {
	E_1 CreateNatGatewayOptionSpec
	E_2 CreateNatGatewayOptionSpec
	E_3 CreateNatGatewayOptionSpec
	E_4 CreateNatGatewayOptionSpec
}

func GetCreateNatGatewayOptionSpecEnum() CreateNatGatewayOptionSpecEnum {
	return CreateNatGatewayOptionSpecEnum{
		E_1: CreateNatGatewayOptionSpec{
			value: "1",
		},
		E_2: CreateNatGatewayOptionSpec{
			value: "2",
		},
		E_3: CreateNatGatewayOptionSpec{
			value: "3",
		},
		E_4: CreateNatGatewayOptionSpec{
			value: "4",
		},
	}
}

func (c CreateNatGatewayOptionSpec) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateNatGatewayOptionSpec) UnmarshalJSON(b []byte) error {
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

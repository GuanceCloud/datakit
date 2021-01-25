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

// 更新公网NAT网关实例的请求体。
type UpdateNatGatewayOption struct {
	// 公网NAT网关实例的名字，长度限制为64。 公网NAT网关实例的名字仅支持数字、字母、_（下划线）、-（中划线）、中文。
	Name *string `json:"name,omitempty"`
	// 公网NAT网关的描述，长度限制为255。
	Description *string `json:"description,omitempty"`
	// 公网NAT网关的规格。 取值为： \"1\"：小型，SNAT最大连接数10000 \"2\"：中型，SNAT最大连接数50000 \"3\"：大型，SNAT最大连接数200000 \"4\"：超大型，SNAT最大连接数1000000
	Spec *UpdateNatGatewayOptionSpec `json:"spec,omitempty"`
}

func (o UpdateNatGatewayOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateNatGatewayOption struct{}"
	}

	return strings.Join([]string{"UpdateNatGatewayOption", string(data)}, " ")
}

type UpdateNatGatewayOptionSpec struct {
	value string
}

type UpdateNatGatewayOptionSpecEnum struct {
	E_1 UpdateNatGatewayOptionSpec
	E_2 UpdateNatGatewayOptionSpec
	E_3 UpdateNatGatewayOptionSpec
	E_4 UpdateNatGatewayOptionSpec
}

func GetUpdateNatGatewayOptionSpecEnum() UpdateNatGatewayOptionSpecEnum {
	return UpdateNatGatewayOptionSpecEnum{
		E_1: UpdateNatGatewayOptionSpec{
			value: "1",
		},
		E_2: UpdateNatGatewayOptionSpec{
			value: "2",
		},
		E_3: UpdateNatGatewayOptionSpec{
			value: "3",
		},
		E_4: UpdateNatGatewayOptionSpec{
			value: "4",
		},
	}
}

func (c UpdateNatGatewayOptionSpec) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdateNatGatewayOptionSpec) UnmarshalJSON(b []byte) error {
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

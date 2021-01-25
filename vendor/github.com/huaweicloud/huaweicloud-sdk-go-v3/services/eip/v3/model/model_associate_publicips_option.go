/*
 * EIP
 *
 * 云服务接口
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

type AssociatePublicipsOption struct {
	// 功能说明：端口所属实例类型 取值范围：PORT、NATGW、VPN、ELB、null 约束：associate_instance_type和associate_instance_id都不为空时表示绑定实例； associate_instance_type和associate_instance_id都为null时解绑实例 约束：双栈公网IP不允许修改绑定的实例
	AssociateInstanceType *AssociatePublicipsOptionAssociateInstanceType `json:"associate_instance_type,omitempty"`
	// 功能说明：端口所属实例ID，例如RDS实例ID 约束：associate_instance_type和associate_instance_id都不为空时表示绑定实例； associate_instance_type和associate_instance_id都为null时解绑实例 约束：双栈公网IP不允许修改绑定的实例
	AssociateInstanceId *string `json:"associate_instance_id,omitempty"`
}

func (o AssociatePublicipsOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AssociatePublicipsOption struct{}"
	}

	return strings.Join([]string{"AssociatePublicipsOption", string(data)}, " ")
}

type AssociatePublicipsOptionAssociateInstanceType struct {
	value string
}

type AssociatePublicipsOptionAssociateInstanceTypeEnum struct {
	PORT  AssociatePublicipsOptionAssociateInstanceType
	NATGW AssociatePublicipsOptionAssociateInstanceType
	ELB   AssociatePublicipsOptionAssociateInstanceType
	EMPTY AssociatePublicipsOptionAssociateInstanceType
}

func GetAssociatePublicipsOptionAssociateInstanceTypeEnum() AssociatePublicipsOptionAssociateInstanceTypeEnum {
	return AssociatePublicipsOptionAssociateInstanceTypeEnum{
		PORT: AssociatePublicipsOptionAssociateInstanceType{
			value: "PORT",
		},
		NATGW: AssociatePublicipsOptionAssociateInstanceType{
			value: "NATGW",
		},
		ELB: AssociatePublicipsOptionAssociateInstanceType{
			value: "ELB",
		},
		EMPTY: AssociatePublicipsOptionAssociateInstanceType{
			value: "",
		},
	}
}

func (c AssociatePublicipsOptionAssociateInstanceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *AssociatePublicipsOptionAssociateInstanceType) UnmarshalJSON(b []byte) error {
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

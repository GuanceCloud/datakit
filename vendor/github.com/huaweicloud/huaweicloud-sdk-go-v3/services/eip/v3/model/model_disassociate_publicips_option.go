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

type DisassociatePublicipsOption struct {
	// 功能说明：端口所属实例类型 取值范围：PORT、NATGW、VPN、ELB、null 约束：associate_instance_type和associate_instance_id都不为空时表示绑定实例； associate_instance_type和associate_instance_id都为null时解绑实例 约束：双栈公网IP不允许修改绑定的实例
	AssociateInstanceType *DisassociatePublicipsOptionAssociateInstanceType `json:"associate_instance_type,omitempty"`
	// 功能说明：端口所属实例ID，例如RDS实例ID 约束：associate_instance_type和associate_instance_id都不为空时表示绑定实例； associate_instance_type和associate_instance_id都为null时解绑实例 约束：双栈公网IP不允许修改绑定的实例
	AssociateInstanceId *string `json:"associate_instance_id,omitempty"`
}

func (o DisassociatePublicipsOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DisassociatePublicipsOption struct{}"
	}

	return strings.Join([]string{"DisassociatePublicipsOption", string(data)}, " ")
}

type DisassociatePublicipsOptionAssociateInstanceType struct {
	value string
}

type DisassociatePublicipsOptionAssociateInstanceTypeEnum struct {
	PORT  DisassociatePublicipsOptionAssociateInstanceType
	NATGW DisassociatePublicipsOptionAssociateInstanceType
	ELB   DisassociatePublicipsOptionAssociateInstanceType
	EMPTY DisassociatePublicipsOptionAssociateInstanceType
}

func GetDisassociatePublicipsOptionAssociateInstanceTypeEnum() DisassociatePublicipsOptionAssociateInstanceTypeEnum {
	return DisassociatePublicipsOptionAssociateInstanceTypeEnum{
		PORT: DisassociatePublicipsOptionAssociateInstanceType{
			value: "PORT",
		},
		NATGW: DisassociatePublicipsOptionAssociateInstanceType{
			value: "NATGW",
		},
		ELB: DisassociatePublicipsOptionAssociateInstanceType{
			value: "ELB",
		},
		EMPTY: DisassociatePublicipsOptionAssociateInstanceType{
			value: "",
		},
	}
}

func (c DisassociatePublicipsOptionAssociateInstanceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DisassociatePublicipsOptionAssociateInstanceType) UnmarshalJSON(b []byte) error {
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

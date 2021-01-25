/*
 * ServiceStage
 *
 * ServiceStage的API,包括应用管理和仓库授权管理
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// 资源类型。  基础资源：cce、cci、ecs、as。  可选资源：rds、dcs、elb等其他类型。
type ResourceType struct {
	value string
}

type ResourceTypeEnum struct {
	SECURITY_GROUP ResourceType
	EIP            ResourceType
	ELB            ResourceType
	CCE            ResourceType
	CCI            ResourceType
	ECS            ResourceType
	AS             ResourceType
	CSE            ResourceType
	DCS            ResourceType
	RDS            ResourceType
	PVC            ResourceType
	APM            ResourceType
}

func GetResourceTypeEnum() ResourceTypeEnum {
	return ResourceTypeEnum{
		SECURITY_GROUP: ResourceType{
			value: "security_group",
		},
		EIP: ResourceType{
			value: "eip",
		},
		ELB: ResourceType{
			value: "elb",
		},
		CCE: ResourceType{
			value: "cce",
		},
		CCI: ResourceType{
			value: "cci",
		},
		ECS: ResourceType{
			value: "ecs",
		},
		AS: ResourceType{
			value: "as",
		},
		CSE: ResourceType{
			value: "cse",
		},
		DCS: ResourceType{
			value: "dcs",
		},
		RDS: ResourceType{
			value: "rds",
		},
		PVC: ResourceType{
			value: "pvc",
		},
		APM: ResourceType{
			value: "apm",
		},
	}
}

func (c ResourceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ResourceType) UnmarshalJSON(b []byte) error {
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

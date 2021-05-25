/*
 * RMS
 *
 * Resource Manager Api
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// 策略参数定义
type PolicyParameterDefinition struct {
	// 策略参数名字
	Name *string `json:"name,omitempty"`
	// 策略参数描述
	Description *string `json:"description,omitempty"`
	// 策略参数允许值列表
	AllowedValues *[]interface{} `json:"allowed_values,omitempty"`
	// 策略参数默认值
	DefaultValue *string `json:"default_value,omitempty"`
	// 策略参数类型
	Type *PolicyParameterDefinitionType `json:"type,omitempty"`
}

func (o PolicyParameterDefinition) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PolicyParameterDefinition struct{}"
	}

	return strings.Join([]string{"PolicyParameterDefinition", string(data)}, " ")
}

type PolicyParameterDefinitionType struct {
	value string
}

type PolicyParameterDefinitionTypeEnum struct {
	ARRAY     PolicyParameterDefinitionType
	BOOLEAN   PolicyParameterDefinitionType
	INTEGER   PolicyParameterDefinitionType
	FLOAT     PolicyParameterDefinitionType
	STRING    PolicyParameterDefinitionType
	DATE_TIME PolicyParameterDefinitionType
}

func GetPolicyParameterDefinitionTypeEnum() PolicyParameterDefinitionTypeEnum {
	return PolicyParameterDefinitionTypeEnum{
		ARRAY: PolicyParameterDefinitionType{
			value: "Array",
		},
		BOOLEAN: PolicyParameterDefinitionType{
			value: "Boolean",
		},
		INTEGER: PolicyParameterDefinitionType{
			value: "Integer",
		},
		FLOAT: PolicyParameterDefinitionType{
			value: "Float",
		},
		STRING: PolicyParameterDefinitionType{
			value: "String",
		},
		DATE_TIME: PolicyParameterDefinitionType{
			value: "DateTime",
		},
	}
}

func (c PolicyParameterDefinitionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *PolicyParameterDefinitionType) UnmarshalJSON(b []byte) error {
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

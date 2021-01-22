/*
 * APIG
 *
 * API网关（API Gateway）是为开发者、合作伙伴提供的高性能、高可用、高安全的API托管服务，帮助用户轻松构建、管理和发布任意规模的API。
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

type ApiConditionBase struct {
	// 关联的请求参数对象名称。策略类型为param时必选
	ReqParamName *string `json:"req_param_name,omitempty"`
	// 策略条件 - exact：绝对匹配 - enum：枚举 - pattern：正则  策略类型为param时必选
	ConditionType *ApiConditionBaseConditionType `json:"condition_type,omitempty"`
	// 策略类型 - param：参数 - source：源IP
	ConditionOrigin ApiConditionBaseConditionOrigin `json:"condition_origin"`
	// 策略值
	ConditionValue string `json:"condition_value"`
}

func (o ApiConditionBase) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ApiConditionBase struct{}"
	}

	return strings.Join([]string{"ApiConditionBase", string(data)}, " ")
}

type ApiConditionBaseConditionType struct {
	value string
}

type ApiConditionBaseConditionTypeEnum struct {
	EXACT   ApiConditionBaseConditionType
	ENUM    ApiConditionBaseConditionType
	PATTERN ApiConditionBaseConditionType
}

func GetApiConditionBaseConditionTypeEnum() ApiConditionBaseConditionTypeEnum {
	return ApiConditionBaseConditionTypeEnum{
		EXACT: ApiConditionBaseConditionType{
			value: "exact",
		},
		ENUM: ApiConditionBaseConditionType{
			value: "enum",
		},
		PATTERN: ApiConditionBaseConditionType{
			value: "pattern",
		},
	}
}

func (c ApiConditionBaseConditionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ApiConditionBaseConditionType) UnmarshalJSON(b []byte) error {
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

type ApiConditionBaseConditionOrigin struct {
	value string
}

type ApiConditionBaseConditionOriginEnum struct {
	PARAM  ApiConditionBaseConditionOrigin
	SOURCE ApiConditionBaseConditionOrigin
}

func GetApiConditionBaseConditionOriginEnum() ApiConditionBaseConditionOriginEnum {
	return ApiConditionBaseConditionOriginEnum{
		PARAM: ApiConditionBaseConditionOrigin{
			value: "param",
		},
		SOURCE: ApiConditionBaseConditionOrigin{
			value: "source",
		},
	}
}

func (c ApiConditionBaseConditionOrigin) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ApiConditionBaseConditionOrigin) UnmarshalJSON(b []byte) error {
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

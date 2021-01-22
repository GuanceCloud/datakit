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

type CoditionResp struct {
	// 关联的请求参数对象名称。策略类型为param时必选
	ReqParamName *string `json:"req_param_name,omitempty"`
	// 策略条件 - exact：绝对匹配 - enum：枚举 - pattern：正则  策略类型为param时必选
	ConditionType *CoditionRespConditionType `json:"condition_type,omitempty"`
	// 策略类型 - param：参数 - source：源IP
	ConditionOrigin CoditionRespConditionOrigin `json:"condition_origin"`
	// 策略值
	ConditionValue string `json:"condition_value"`
	// 编号
	Id *string `json:"id,omitempty"`
	// 关联的请求参数对象编号
	ReqParamId *string `json:"req_param_id,omitempty"`
	// 关联的请求参数对象位置
	ReqParamLocation *string `json:"req_param_location,omitempty"`
}

func (o CoditionResp) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CoditionResp struct{}"
	}

	return strings.Join([]string{"CoditionResp", string(data)}, " ")
}

type CoditionRespConditionType struct {
	value string
}

type CoditionRespConditionTypeEnum struct {
	EXACT   CoditionRespConditionType
	ENUM    CoditionRespConditionType
	PATTERN CoditionRespConditionType
}

func GetCoditionRespConditionTypeEnum() CoditionRespConditionTypeEnum {
	return CoditionRespConditionTypeEnum{
		EXACT: CoditionRespConditionType{
			value: "exact",
		},
		ENUM: CoditionRespConditionType{
			value: "enum",
		},
		PATTERN: CoditionRespConditionType{
			value: "pattern",
		},
	}
}

func (c CoditionRespConditionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CoditionRespConditionType) UnmarshalJSON(b []byte) error {
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

type CoditionRespConditionOrigin struct {
	value string
}

type CoditionRespConditionOriginEnum struct {
	PARAM  CoditionRespConditionOrigin
	SOURCE CoditionRespConditionOrigin
}

func GetCoditionRespConditionOriginEnum() CoditionRespConditionOriginEnum {
	return CoditionRespConditionOriginEnum{
		PARAM: CoditionRespConditionOrigin{
			value: "param",
		},
		SOURCE: CoditionRespConditionOrigin{
			value: "source",
		},
	}
}

func (c CoditionRespConditionOrigin) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CoditionRespConditionOrigin) UnmarshalJSON(b []byte) error {
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

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

type ApiPolicyFunctionCreate struct {
	// 函数URN
	FunctionUrn string `json:"function_urn"`
	// 调用类型 - async： 异步 - sync：同步
	InvocationType ApiPolicyFunctionCreateInvocationType `json:"invocation_type"`
	// 版本。字符长度不超过64
	Version *string `json:"version,omitempty"`
	// API网关请求后端服务的超时时间。  单位：毫秒。请求参数值不在合法范围内时将使用默认值
	Timeout *int32 `json:"timeout,omitempty"`
	// 关联的策略组合模式： - ALL：满足全部条件 - ANY：满足任一条件
	EffectMode ApiPolicyFunctionCreateEffectMode `json:"effect_mode"`
	// 策略后端名称。字符串由中文、英文字母、数字、下划线组成，且只能以中文或英文开头。
	Name string `json:"name"`
	// 后端参数列表
	BackendParams *[]BackendParamBase `json:"backend_params,omitempty"`
	// 策略条件列表
	Conditions []ApiConditionBase `json:"conditions"`
	// 后端自定义认证对象的ID
	AuthorizerId *string `json:"authorizer_id,omitempty"`
}

func (o ApiPolicyFunctionCreate) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ApiPolicyFunctionCreate struct{}"
	}

	return strings.Join([]string{"ApiPolicyFunctionCreate", string(data)}, " ")
}

type ApiPolicyFunctionCreateInvocationType struct {
	value string
}

type ApiPolicyFunctionCreateInvocationTypeEnum struct {
	ASYNC ApiPolicyFunctionCreateInvocationType
	SYNC  ApiPolicyFunctionCreateInvocationType
}

func GetApiPolicyFunctionCreateInvocationTypeEnum() ApiPolicyFunctionCreateInvocationTypeEnum {
	return ApiPolicyFunctionCreateInvocationTypeEnum{
		ASYNC: ApiPolicyFunctionCreateInvocationType{
			value: "async",
		},
		SYNC: ApiPolicyFunctionCreateInvocationType{
			value: "sync",
		},
	}
}

func (c ApiPolicyFunctionCreateInvocationType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ApiPolicyFunctionCreateInvocationType) UnmarshalJSON(b []byte) error {
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

type ApiPolicyFunctionCreateEffectMode struct {
	value string
}

type ApiPolicyFunctionCreateEffectModeEnum struct {
	ALL ApiPolicyFunctionCreateEffectMode
	ANY ApiPolicyFunctionCreateEffectMode
}

func GetApiPolicyFunctionCreateEffectModeEnum() ApiPolicyFunctionCreateEffectModeEnum {
	return ApiPolicyFunctionCreateEffectModeEnum{
		ALL: ApiPolicyFunctionCreateEffectMode{
			value: "ALL",
		},
		ANY: ApiPolicyFunctionCreateEffectMode{
			value: "ANY",
		},
	}
}

func (c ApiPolicyFunctionCreateEffectMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ApiPolicyFunctionCreateEffectMode) UnmarshalJSON(b []byte) error {
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

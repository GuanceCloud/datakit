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

type ApiPolicyFunctionBase struct {
	// 函数URN
	FunctionUrn string `json:"function_urn"`
	// 调用类型 - async： 异步 - sync：同步
	InvocationType ApiPolicyFunctionBaseInvocationType `json:"invocation_type"`
	// 版本。字符长度不超过64
	Version *string `json:"version,omitempty"`
	// API网关请求后端服务的超时时间。  单位：毫秒。请求参数值不在合法范围内时将使用默认值
	Timeout *int32 `json:"timeout,omitempty"`
}

func (o ApiPolicyFunctionBase) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ApiPolicyFunctionBase struct{}"
	}

	return strings.Join([]string{"ApiPolicyFunctionBase", string(data)}, " ")
}

type ApiPolicyFunctionBaseInvocationType struct {
	value string
}

type ApiPolicyFunctionBaseInvocationTypeEnum struct {
	ASYNC ApiPolicyFunctionBaseInvocationType
	SYNC  ApiPolicyFunctionBaseInvocationType
}

func GetApiPolicyFunctionBaseInvocationTypeEnum() ApiPolicyFunctionBaseInvocationTypeEnum {
	return ApiPolicyFunctionBaseInvocationTypeEnum{
		ASYNC: ApiPolicyFunctionBaseInvocationType{
			value: "async",
		},
		SYNC: ApiPolicyFunctionBaseInvocationType{
			value: "sync",
		},
	}
}

func (c ApiPolicyFunctionBaseInvocationType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ApiPolicyFunctionBaseInvocationType) UnmarshalJSON(b []byte) error {
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

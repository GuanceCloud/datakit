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

// 函数后端详情
type ApiFuncCreate struct {
	// 函数URN
	FunctionUrn string `json:"function_urn"`
	// 描述信息。长度不超过255个字符 > 中文字符必须为UTF-8或者unicode编码。
	Remark *string `json:"remark,omitempty"`
	// 调用类型 - async： 异步 - sync：同步
	InvocationType ApiFuncCreateInvocationType `json:"invocation_type"`
	// 版本。
	Version *string `json:"version,omitempty"`
	// API网关请求函数服务的超时时间。  单位：毫秒。请求参数值不在合法范围内时将使用缺省值
	Timeout int32 `json:"timeout"`
	// 后端自定义认证ID
	AuthorizerId *string `json:"authorizer_id,omitempty"`
}

func (o ApiFuncCreate) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ApiFuncCreate struct{}"
	}

	return strings.Join([]string{"ApiFuncCreate", string(data)}, " ")
}

type ApiFuncCreateInvocationType struct {
	value string
}

type ApiFuncCreateInvocationTypeEnum struct {
	ASYNC ApiFuncCreateInvocationType
	SYNC  ApiFuncCreateInvocationType
}

func GetApiFuncCreateInvocationTypeEnum() ApiFuncCreateInvocationTypeEnum {
	return ApiFuncCreateInvocationTypeEnum{
		ASYNC: ApiFuncCreateInvocationType{
			value: "async",
		},
		SYNC: ApiFuncCreateInvocationType{
			value: "sync",
		},
	}
}

func (c ApiFuncCreateInvocationType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ApiFuncCreateInvocationType) UnmarshalJSON(b []byte) error {
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

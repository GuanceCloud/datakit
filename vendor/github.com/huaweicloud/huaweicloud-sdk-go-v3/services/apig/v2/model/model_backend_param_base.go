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

type BackendParamBase struct {
	// 参数类别：REQUEST、CONSTANT、SYSTEM
	Origin BackendParamBaseOrigin `json:"origin"`
	// 参数名称。 字符串由英文字母、数字、中划线、下划线、英文句号组成，且只能以英文开头。
	Name string `json:"name"`
	// 描述。字符长度不超过255 > 中文字符必须为UTF-8或者unicode编码。
	Remark *string `json:"remark,omitempty"`
	// 参数位置：PATH、QUERY、HEADER
	Location BackendParamBaseLocation `json:"location"`
	// 参数值。字符长度不超过255，类别为REQUEST时，值为req_params中的参数名称；类别为CONSTANT时，值为参数真正的值；类别为SYSTEM时，值为网关参数名称
	Value string `json:"value"`
}

func (o BackendParamBase) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BackendParamBase struct{}"
	}

	return strings.Join([]string{"BackendParamBase", string(data)}, " ")
}

type BackendParamBaseOrigin struct {
	value string
}

type BackendParamBaseOriginEnum struct {
	REQUEST  BackendParamBaseOrigin
	CONSTANT BackendParamBaseOrigin
	SYSTEM   BackendParamBaseOrigin
}

func GetBackendParamBaseOriginEnum() BackendParamBaseOriginEnum {
	return BackendParamBaseOriginEnum{
		REQUEST: BackendParamBaseOrigin{
			value: "REQUEST",
		},
		CONSTANT: BackendParamBaseOrigin{
			value: "CONSTANT",
		},
		SYSTEM: BackendParamBaseOrigin{
			value: "SYSTEM",
		},
	}
}

func (c BackendParamBaseOrigin) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BackendParamBaseOrigin) UnmarshalJSON(b []byte) error {
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

type BackendParamBaseLocation struct {
	value string
}

type BackendParamBaseLocationEnum struct {
	PATH   BackendParamBaseLocation
	QUERY  BackendParamBaseLocation
	HEADER BackendParamBaseLocation
}

func GetBackendParamBaseLocationEnum() BackendParamBaseLocationEnum {
	return BackendParamBaseLocationEnum{
		PATH: BackendParamBaseLocation{
			value: "PATH",
		},
		QUERY: BackendParamBaseLocation{
			value: "QUERY",
		},
		HEADER: BackendParamBaseLocation{
			value: "HEADER",
		},
	}
}

func (c BackendParamBaseLocation) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BackendParamBaseLocation) UnmarshalJSON(b []byte) error {
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

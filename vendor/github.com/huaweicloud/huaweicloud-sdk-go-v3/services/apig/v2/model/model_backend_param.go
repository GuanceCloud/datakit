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

type BackendParam struct {
	// 参数类别：REQUEST、CONSTANT、SYSTEM
	Origin BackendParamOrigin `json:"origin"`
	// 参数名称。 字符串由英文字母、数字、中划线、下划线、英文句号组成，且只能以英文开头。
	Name string `json:"name"`
	// 描述。字符长度不超过255 > 中文字符必须为UTF-8或者unicode编码。
	Remark *string `json:"remark,omitempty"`
	// 参数位置：PATH、QUERY、HEADER
	Location BackendParamLocation `json:"location"`
	// 参数值。字符长度不超过255，类别为REQUEST时，值为req_params中的参数名称；类别为CONSTANT时，值为参数真正的值；类别为SYSTEM时，值为网关参数名称
	Value string `json:"value"`
	// 参数编号
	Id *string `json:"id,omitempty"`
	// 对应的请求参数编号
	ReqParamId *string `json:"req_param_id,omitempty"`
}

func (o BackendParam) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BackendParam struct{}"
	}

	return strings.Join([]string{"BackendParam", string(data)}, " ")
}

type BackendParamOrigin struct {
	value string
}

type BackendParamOriginEnum struct {
	REQUEST  BackendParamOrigin
	CONSTANT BackendParamOrigin
	SYSTEM   BackendParamOrigin
}

func GetBackendParamOriginEnum() BackendParamOriginEnum {
	return BackendParamOriginEnum{
		REQUEST: BackendParamOrigin{
			value: "REQUEST",
		},
		CONSTANT: BackendParamOrigin{
			value: "CONSTANT",
		},
		SYSTEM: BackendParamOrigin{
			value: "SYSTEM",
		},
	}
}

func (c BackendParamOrigin) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BackendParamOrigin) UnmarshalJSON(b []byte) error {
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

type BackendParamLocation struct {
	value string
}

type BackendParamLocationEnum struct {
	PATH   BackendParamLocation
	QUERY  BackendParamLocation
	HEADER BackendParamLocation
}

func GetBackendParamLocationEnum() BackendParamLocationEnum {
	return BackendParamLocationEnum{
		PATH: BackendParamLocation{
			value: "PATH",
		},
		QUERY: BackendParamLocation{
			value: "QUERY",
		},
		HEADER: BackendParamLocation{
			value: "HEADER",
		},
	}
}

func (c BackendParamLocation) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BackendParamLocation) UnmarshalJSON(b []byte) error {
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

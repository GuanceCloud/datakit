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
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"
	"strings"
)

type BackendApiBase struct {
	// 后端自定义认证对象的ID
	AuthorizerId *string `json:"authorizer_id,omitempty"`
	// 后端服务的地址。  由主机（IP或域名）和端口号组成，总长度不超过255。格式为主机:端口（如：apig.example.com:7443）。如果不写端口，则HTTPS默认端口号为443，HTTP默认端口号为80。  支持环境变量，使用环境变量时，每个变量名的长度为3 ~ 32位的字符串，字符串由英文字母、数字、下划线、中划线组成，且只能以英文开头
	UrlDomain *string `json:"url_domain,omitempty"`
	// 请求协议
	ReqProtocol BackendApiBaseReqProtocol `json:"req_protocol"`
	// 描述。字符长度不超过255 > 中文字符必须为UTF-8或者unicode编码。
	Remark *string `json:"remark,omitempty"`
	// 请求方式
	ReqMethod BackendApiBaseReqMethod `json:"req_method"`
	// web后端版本，字符长度不超过16
	Version *string `json:"version,omitempty"`
	// 请求地址。可以包含请求参数，用{}标识，比如/getUserInfo/{userId}，支持 * % - _ . 等特殊字符，总长度不超过512，且满足URI规范。  支持环境变量，使用环境变量时，每个变量名的长度为3 ~ 32位的字符串，字符串由英文字母、数字、中划线、下划线组成，且只能以英文开头。 > 需要服从URI规范。
	ReqUri string `json:"req_uri"`
	// API网关请求后端服务的超时时间。  单位：毫秒。请求参数值不在合法范围内时将使用默认值
	Timeout int32 `json:"timeout"`
	// 编号
	Id *string `json:"id,omitempty"`
	// 状态
	Status *int32 `json:"status,omitempty"`
	// 注册时间
	RegisterTime *sdktime.SdkTime `json:"register_time,omitempty"`
	// 修改时间
	UpdateTime *sdktime.SdkTime `json:"update_time,omitempty"`
}

func (o BackendApiBase) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BackendApiBase struct{}"
	}

	return strings.Join([]string{"BackendApiBase", string(data)}, " ")
}

type BackendApiBaseReqProtocol struct {
	value string
}

type BackendApiBaseReqProtocolEnum struct {
	HTTP  BackendApiBaseReqProtocol
	HTTPS BackendApiBaseReqProtocol
}

func GetBackendApiBaseReqProtocolEnum() BackendApiBaseReqProtocolEnum {
	return BackendApiBaseReqProtocolEnum{
		HTTP: BackendApiBaseReqProtocol{
			value: "HTTP",
		},
		HTTPS: BackendApiBaseReqProtocol{
			value: "HTTPS",
		},
	}
}

func (c BackendApiBaseReqProtocol) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BackendApiBaseReqProtocol) UnmarshalJSON(b []byte) error {
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

type BackendApiBaseReqMethod struct {
	value string
}

type BackendApiBaseReqMethodEnum struct {
	GET     BackendApiBaseReqMethod
	POST    BackendApiBaseReqMethod
	PUT     BackendApiBaseReqMethod
	DELETE  BackendApiBaseReqMethod
	HEAD    BackendApiBaseReqMethod
	PATCH   BackendApiBaseReqMethod
	OPTIONS BackendApiBaseReqMethod
	ANY     BackendApiBaseReqMethod
}

func GetBackendApiBaseReqMethodEnum() BackendApiBaseReqMethodEnum {
	return BackendApiBaseReqMethodEnum{
		GET: BackendApiBaseReqMethod{
			value: "GET",
		},
		POST: BackendApiBaseReqMethod{
			value: "POST",
		},
		PUT: BackendApiBaseReqMethod{
			value: "PUT",
		},
		DELETE: BackendApiBaseReqMethod{
			value: "DELETE",
		},
		HEAD: BackendApiBaseReqMethod{
			value: "HEAD",
		},
		PATCH: BackendApiBaseReqMethod{
			value: "PATCH",
		},
		OPTIONS: BackendApiBaseReqMethod{
			value: "OPTIONS",
		},
		ANY: BackendApiBaseReqMethod{
			value: "ANY",
		},
	}
}

func (c BackendApiBaseReqMethod) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BackendApiBaseReqMethod) UnmarshalJSON(b []byte) error {
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

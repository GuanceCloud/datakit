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

type ApiPolicyHttpBase struct {
	// 策略后端的Endpoint。 由域名（或IP地址）和端口号组成，总长度不超过255。格式为域名:端口（如：apig.example.com:7443）。如果不写端口，则HTTPS默认端口号为443， HTTP默认端口号为80。 支持环境变量，使用环境变量时，每个变量名的长度为3 ~ 32位的字符串，字符串由英文字母、数字、“_”、“-”组成，且只能以英文开头。
	UrlDomain *string `json:"url_domain,omitempty"`
	// 请求协议：HTTP、HTTPS
	ReqProtocol ApiPolicyHttpBaseReqProtocol `json:"req_protocol"`
	// 请求方式：GET、POST、PUT、DELETE、HEAD、PATCH、OPTIONS、ANY
	ReqMethod ApiPolicyHttpBaseReqMethod `json:"req_method"`
	// 请求地址。可以包含请求参数，用{}标识，比如/getUserInfo/{userId}，支持 * % - _ . 等特殊字符，总长度不超过512，且满足URI规范。  支持环境变量，使用环境变量时，每个变量名的长度为3 ~ 32位的字符串，字符串由英文字母、数字、中划线、下划线组成，且只能以英文开头。 > 需要服从URI规范。
	ReqUri string `json:"req_uri"`
	// API网关请求后端服务的超时时间。  单位：毫秒。
	Timeout *int32 `json:"timeout,omitempty"`
}

func (o ApiPolicyHttpBase) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ApiPolicyHttpBase struct{}"
	}

	return strings.Join([]string{"ApiPolicyHttpBase", string(data)}, " ")
}

type ApiPolicyHttpBaseReqProtocol struct {
	value string
}

type ApiPolicyHttpBaseReqProtocolEnum struct {
	HTTP  ApiPolicyHttpBaseReqProtocol
	HTTPS ApiPolicyHttpBaseReqProtocol
}

func GetApiPolicyHttpBaseReqProtocolEnum() ApiPolicyHttpBaseReqProtocolEnum {
	return ApiPolicyHttpBaseReqProtocolEnum{
		HTTP: ApiPolicyHttpBaseReqProtocol{
			value: "HTTP",
		},
		HTTPS: ApiPolicyHttpBaseReqProtocol{
			value: "HTTPS",
		},
	}
}

func (c ApiPolicyHttpBaseReqProtocol) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ApiPolicyHttpBaseReqProtocol) UnmarshalJSON(b []byte) error {
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

type ApiPolicyHttpBaseReqMethod struct {
	value string
}

type ApiPolicyHttpBaseReqMethodEnum struct {
	GET     ApiPolicyHttpBaseReqMethod
	POST    ApiPolicyHttpBaseReqMethod
	PUT     ApiPolicyHttpBaseReqMethod
	DELETE  ApiPolicyHttpBaseReqMethod
	HEAD    ApiPolicyHttpBaseReqMethod
	PATCH   ApiPolicyHttpBaseReqMethod
	OPTIONS ApiPolicyHttpBaseReqMethod
	ANY     ApiPolicyHttpBaseReqMethod
}

func GetApiPolicyHttpBaseReqMethodEnum() ApiPolicyHttpBaseReqMethodEnum {
	return ApiPolicyHttpBaseReqMethodEnum{
		GET: ApiPolicyHttpBaseReqMethod{
			value: "GET",
		},
		POST: ApiPolicyHttpBaseReqMethod{
			value: "POST",
		},
		PUT: ApiPolicyHttpBaseReqMethod{
			value: "PUT",
		},
		DELETE: ApiPolicyHttpBaseReqMethod{
			value: "DELETE",
		},
		HEAD: ApiPolicyHttpBaseReqMethod{
			value: "HEAD",
		},
		PATCH: ApiPolicyHttpBaseReqMethod{
			value: "PATCH",
		},
		OPTIONS: ApiPolicyHttpBaseReqMethod{
			value: "OPTIONS",
		},
		ANY: ApiPolicyHttpBaseReqMethod{
			value: "ANY",
		},
	}
}

func (c ApiPolicyHttpBaseReqMethod) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ApiPolicyHttpBaseReqMethod) UnmarshalJSON(b []byte) error {
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

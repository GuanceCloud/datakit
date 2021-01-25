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

type ApiPolicyHttpResp struct {
	// 编号
	Id *string `json:"id,omitempty"`
	// 关联的策略组合模式： - ALL：满足全部条件 - ANY：满足任一条件
	EffectMode ApiPolicyHttpRespEffectMode `json:"effect_mode"`
	// 策略后端名称。字符串由中文、英文字母、数字、下划线组成，且只能以中文或英文开头。
	Name string `json:"name"`
	// 后端参数列表
	BackendParams *[]BackendParam `json:"backend_params,omitempty"`
	// 策略条件列表
	Conditions []CoditionResp `json:"conditions"`
	// 后端自定义认证对象的ID
	AuthorizerId *string `json:"authorizer_id,omitempty"`
	// 策略后端的Endpoint。 由域名（或IP地址）和端口号组成，总长度不超过255。格式为域名:端口（如：apig.example.com:7443）。如果不写端口，则HTTPS默认端口号为443， HTTP默认端口号为80。 支持环境变量，使用环境变量时，每个变量名的长度为3 ~ 32位的字符串，字符串由英文字母、数字、“_”、“-”组成，且只能以英文开头。
	UrlDomain *string `json:"url_domain,omitempty"`
	// 请求协议：HTTP、HTTPS
	ReqProtocol ApiPolicyHttpRespReqProtocol `json:"req_protocol"`
	// 请求方式：GET、POST、PUT、DELETE、HEAD、PATCH、OPTIONS、ANY
	ReqMethod ApiPolicyHttpRespReqMethod `json:"req_method"`
	// 请求地址。可以包含请求参数，用{}标识，比如/getUserInfo/{userId}，支持 * % - _ . 等特殊字符，总长度不超过512，且满足URI规范。  支持环境变量，使用环境变量时，每个变量名的长度为3 ~ 32位的字符串，字符串由英文字母、数字、中划线、下划线组成，且只能以英文开头。 > 需要服从URI规范。
	ReqUri string `json:"req_uri"`
	// API网关请求后端服务的超时时间。  单位：毫秒。
	Timeout *int32 `json:"timeout,omitempty"`
	// VPC通道信息
	VpcChannelInfo *string `json:"vpc_channel_info,omitempty"`
	// 是否使用VPC通道： - 1： 使用VPC通道 - 2：不使用VPC通道
	VpcChannelStatus *int32 `json:"vpc_channel_status,omitempty"`
}

func (o ApiPolicyHttpResp) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ApiPolicyHttpResp struct{}"
	}

	return strings.Join([]string{"ApiPolicyHttpResp", string(data)}, " ")
}

type ApiPolicyHttpRespEffectMode struct {
	value string
}

type ApiPolicyHttpRespEffectModeEnum struct {
	ALL ApiPolicyHttpRespEffectMode
	ANY ApiPolicyHttpRespEffectMode
}

func GetApiPolicyHttpRespEffectModeEnum() ApiPolicyHttpRespEffectModeEnum {
	return ApiPolicyHttpRespEffectModeEnum{
		ALL: ApiPolicyHttpRespEffectMode{
			value: "ALL",
		},
		ANY: ApiPolicyHttpRespEffectMode{
			value: "ANY",
		},
	}
}

func (c ApiPolicyHttpRespEffectMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ApiPolicyHttpRespEffectMode) UnmarshalJSON(b []byte) error {
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

type ApiPolicyHttpRespReqProtocol struct {
	value string
}

type ApiPolicyHttpRespReqProtocolEnum struct {
	HTTP  ApiPolicyHttpRespReqProtocol
	HTTPS ApiPolicyHttpRespReqProtocol
}

func GetApiPolicyHttpRespReqProtocolEnum() ApiPolicyHttpRespReqProtocolEnum {
	return ApiPolicyHttpRespReqProtocolEnum{
		HTTP: ApiPolicyHttpRespReqProtocol{
			value: "HTTP",
		},
		HTTPS: ApiPolicyHttpRespReqProtocol{
			value: "HTTPS",
		},
	}
}

func (c ApiPolicyHttpRespReqProtocol) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ApiPolicyHttpRespReqProtocol) UnmarshalJSON(b []byte) error {
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

type ApiPolicyHttpRespReqMethod struct {
	value string
}

type ApiPolicyHttpRespReqMethodEnum struct {
	GET     ApiPolicyHttpRespReqMethod
	POST    ApiPolicyHttpRespReqMethod
	PUT     ApiPolicyHttpRespReqMethod
	DELETE  ApiPolicyHttpRespReqMethod
	HEAD    ApiPolicyHttpRespReqMethod
	PATCH   ApiPolicyHttpRespReqMethod
	OPTIONS ApiPolicyHttpRespReqMethod
	ANY     ApiPolicyHttpRespReqMethod
}

func GetApiPolicyHttpRespReqMethodEnum() ApiPolicyHttpRespReqMethodEnum {
	return ApiPolicyHttpRespReqMethodEnum{
		GET: ApiPolicyHttpRespReqMethod{
			value: "GET",
		},
		POST: ApiPolicyHttpRespReqMethod{
			value: "POST",
		},
		PUT: ApiPolicyHttpRespReqMethod{
			value: "PUT",
		},
		DELETE: ApiPolicyHttpRespReqMethod{
			value: "DELETE",
		},
		HEAD: ApiPolicyHttpRespReqMethod{
			value: "HEAD",
		},
		PATCH: ApiPolicyHttpRespReqMethod{
			value: "PATCH",
		},
		OPTIONS: ApiPolicyHttpRespReqMethod{
			value: "OPTIONS",
		},
		ANY: ApiPolicyHttpRespReqMethod{
			value: "ANY",
		},
	}
}

func (c ApiPolicyHttpRespReqMethod) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ApiPolicyHttpRespReqMethod) UnmarshalJSON(b []byte) error {
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

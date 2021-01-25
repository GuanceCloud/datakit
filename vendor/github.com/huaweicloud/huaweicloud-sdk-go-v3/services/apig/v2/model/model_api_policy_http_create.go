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

type ApiPolicyHttpCreate struct {
	// 策略后端的Endpoint。 由域名（或IP地址）和端口号组成，总长度不超过255。格式为域名:端口（如：apig.example.com:7443）。如果不写端口，则HTTPS默认端口号为443， HTTP默认端口号为80。 支持环境变量，使用环境变量时，每个变量名的长度为3 ~ 32位的字符串，字符串由英文字母、数字、“_”、“-”组成，且只能以英文开头。
	UrlDomain *string `json:"url_domain,omitempty"`
	// 请求协议：HTTP、HTTPS
	ReqProtocol ApiPolicyHttpCreateReqProtocol `json:"req_protocol"`
	// 请求方式：GET、POST、PUT、DELETE、HEAD、PATCH、OPTIONS、ANY
	ReqMethod ApiPolicyHttpCreateReqMethod `json:"req_method"`
	// 请求地址。可以包含请求参数，用{}标识，比如/getUserInfo/{userId}，支持 * % - _ . 等特殊字符，总长度不超过512，且满足URI规范。  支持环境变量，使用环境变量时，每个变量名的长度为3 ~ 32位的字符串，字符串由英文字母、数字、中划线、下划线组成，且只能以英文开头。 > 需要服从URI规范。
	ReqUri string `json:"req_uri"`
	// API网关请求后端服务的超时时间。  单位：毫秒。
	Timeout *int32 `json:"timeout,omitempty"`
	// 关联的策略组合模式： - ALL：满足全部条件 - ANY：满足任一条件
	EffectMode ApiPolicyHttpCreateEffectMode `json:"effect_mode"`
	// 策略后端名称。字符串由中文、英文字母、数字、下划线组成，且只能以中文或英文开头。
	Name string `json:"name"`
	// 后端参数列表
	BackendParams *[]BackendParamBase `json:"backend_params,omitempty"`
	// 策略条件列表
	Conditions []ApiConditionBase `json:"conditions"`
	// 后端自定义认证对象的ID
	AuthorizerId   *string           `json:"authorizer_id,omitempty"`
	VpcChannelInfo *ApiBackendVpcReq `json:"vpc_channel_info,omitempty"`
	// 是否使用VPC通道 - 1 : 使用VPC通道 - 2 : 不使用VPC通道
	VpcChannelStatus *ApiPolicyHttpCreateVpcChannelStatus `json:"vpc_channel_status,omitempty"`
}

func (o ApiPolicyHttpCreate) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ApiPolicyHttpCreate struct{}"
	}

	return strings.Join([]string{"ApiPolicyHttpCreate", string(data)}, " ")
}

type ApiPolicyHttpCreateReqProtocol struct {
	value string
}

type ApiPolicyHttpCreateReqProtocolEnum struct {
	HTTP  ApiPolicyHttpCreateReqProtocol
	HTTPS ApiPolicyHttpCreateReqProtocol
}

func GetApiPolicyHttpCreateReqProtocolEnum() ApiPolicyHttpCreateReqProtocolEnum {
	return ApiPolicyHttpCreateReqProtocolEnum{
		HTTP: ApiPolicyHttpCreateReqProtocol{
			value: "HTTP",
		},
		HTTPS: ApiPolicyHttpCreateReqProtocol{
			value: "HTTPS",
		},
	}
}

func (c ApiPolicyHttpCreateReqProtocol) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ApiPolicyHttpCreateReqProtocol) UnmarshalJSON(b []byte) error {
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

type ApiPolicyHttpCreateReqMethod struct {
	value string
}

type ApiPolicyHttpCreateReqMethodEnum struct {
	GET     ApiPolicyHttpCreateReqMethod
	POST    ApiPolicyHttpCreateReqMethod
	PUT     ApiPolicyHttpCreateReqMethod
	DELETE  ApiPolicyHttpCreateReqMethod
	HEAD    ApiPolicyHttpCreateReqMethod
	PATCH   ApiPolicyHttpCreateReqMethod
	OPTIONS ApiPolicyHttpCreateReqMethod
	ANY     ApiPolicyHttpCreateReqMethod
}

func GetApiPolicyHttpCreateReqMethodEnum() ApiPolicyHttpCreateReqMethodEnum {
	return ApiPolicyHttpCreateReqMethodEnum{
		GET: ApiPolicyHttpCreateReqMethod{
			value: "GET",
		},
		POST: ApiPolicyHttpCreateReqMethod{
			value: "POST",
		},
		PUT: ApiPolicyHttpCreateReqMethod{
			value: "PUT",
		},
		DELETE: ApiPolicyHttpCreateReqMethod{
			value: "DELETE",
		},
		HEAD: ApiPolicyHttpCreateReqMethod{
			value: "HEAD",
		},
		PATCH: ApiPolicyHttpCreateReqMethod{
			value: "PATCH",
		},
		OPTIONS: ApiPolicyHttpCreateReqMethod{
			value: "OPTIONS",
		},
		ANY: ApiPolicyHttpCreateReqMethod{
			value: "ANY",
		},
	}
}

func (c ApiPolicyHttpCreateReqMethod) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ApiPolicyHttpCreateReqMethod) UnmarshalJSON(b []byte) error {
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

type ApiPolicyHttpCreateEffectMode struct {
	value string
}

type ApiPolicyHttpCreateEffectModeEnum struct {
	ALL ApiPolicyHttpCreateEffectMode
	ANY ApiPolicyHttpCreateEffectMode
}

func GetApiPolicyHttpCreateEffectModeEnum() ApiPolicyHttpCreateEffectModeEnum {
	return ApiPolicyHttpCreateEffectModeEnum{
		ALL: ApiPolicyHttpCreateEffectMode{
			value: "ALL",
		},
		ANY: ApiPolicyHttpCreateEffectMode{
			value: "ANY",
		},
	}
}

func (c ApiPolicyHttpCreateEffectMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ApiPolicyHttpCreateEffectMode) UnmarshalJSON(b []byte) error {
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

type ApiPolicyHttpCreateVpcChannelStatus struct {
	value int32
}

type ApiPolicyHttpCreateVpcChannelStatusEnum struct {
	E_1 ApiPolicyHttpCreateVpcChannelStatus
	E_2 ApiPolicyHttpCreateVpcChannelStatus
}

func GetApiPolicyHttpCreateVpcChannelStatusEnum() ApiPolicyHttpCreateVpcChannelStatusEnum {
	return ApiPolicyHttpCreateVpcChannelStatusEnum{
		E_1: ApiPolicyHttpCreateVpcChannelStatus{
			value: 1,
		}, E_2: ApiPolicyHttpCreateVpcChannelStatus{
			value: 2,
		},
	}
}

func (c ApiPolicyHttpCreateVpcChannelStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ApiPolicyHttpCreateVpcChannelStatus) UnmarshalJSON(b []byte) error {
	myConverter := converter.StringConverterFactory("int32")
	if myConverter != nil {
		val, err := myConverter.CovertStringToInterface(strings.Trim(string(b[:]), "\""))
		if err == nil {
			c.value = val.(int32)
			return nil
		}
		return err
	} else {
		return errors.New("convert enum data to int32 error")
	}
}

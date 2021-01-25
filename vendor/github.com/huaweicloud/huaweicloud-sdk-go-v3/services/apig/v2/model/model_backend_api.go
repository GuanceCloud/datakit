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

// web后端详情
type BackendApi struct {
	// 后端自定义认证对象的ID
	AuthorizerId *string `json:"authorizer_id,omitempty"`
	// 后端服务的地址。  由主机（IP或域名）和端口号组成，总长度不超过255。格式为主机:端口（如：apig.example.com:7443）。如果不写端口，则HTTPS默认端口号为443，HTTP默认端口号为80。  支持环境变量，使用环境变量时，每个变量名的长度为3 ~ 32位的字符串，字符串由英文字母、数字、下划线、中划线组成，且只能以英文开头
	UrlDomain *string `json:"url_domain,omitempty"`
	// 请求协议
	ReqProtocol BackendApiReqProtocol `json:"req_protocol"`
	// 描述。字符长度不超过255 > 中文字符必须为UTF-8或者unicode编码。
	Remark *string `json:"remark,omitempty"`
	// 请求方式
	ReqMethod BackendApiReqMethod `json:"req_method"`
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
	UpdateTime     *sdktime.SdkTime `json:"update_time,omitempty"`
	VpcChannelInfo *VpcInfo         `json:"vpc_channel_info,omitempty"`
	// 是否使用VPC通道 - 1：使用VPC通道 - 2：不使用VPC通道
	VpcChannelStatus *int32 `json:"vpc_channel_status,omitempty"`
}

func (o BackendApi) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BackendApi struct{}"
	}

	return strings.Join([]string{"BackendApi", string(data)}, " ")
}

type BackendApiReqProtocol struct {
	value string
}

type BackendApiReqProtocolEnum struct {
	HTTP  BackendApiReqProtocol
	HTTPS BackendApiReqProtocol
}

func GetBackendApiReqProtocolEnum() BackendApiReqProtocolEnum {
	return BackendApiReqProtocolEnum{
		HTTP: BackendApiReqProtocol{
			value: "HTTP",
		},
		HTTPS: BackendApiReqProtocol{
			value: "HTTPS",
		},
	}
}

func (c BackendApiReqProtocol) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BackendApiReqProtocol) UnmarshalJSON(b []byte) error {
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

type BackendApiReqMethod struct {
	value string
}

type BackendApiReqMethodEnum struct {
	GET     BackendApiReqMethod
	POST    BackendApiReqMethod
	PUT     BackendApiReqMethod
	DELETE  BackendApiReqMethod
	HEAD    BackendApiReqMethod
	PATCH   BackendApiReqMethod
	OPTIONS BackendApiReqMethod
	ANY     BackendApiReqMethod
}

func GetBackendApiReqMethodEnum() BackendApiReqMethodEnum {
	return BackendApiReqMethodEnum{
		GET: BackendApiReqMethod{
			value: "GET",
		},
		POST: BackendApiReqMethod{
			value: "POST",
		},
		PUT: BackendApiReqMethod{
			value: "PUT",
		},
		DELETE: BackendApiReqMethod{
			value: "DELETE",
		},
		HEAD: BackendApiReqMethod{
			value: "HEAD",
		},
		PATCH: BackendApiReqMethod{
			value: "PATCH",
		},
		OPTIONS: BackendApiReqMethod{
			value: "OPTIONS",
		},
		ANY: BackendApiReqMethod{
			value: "ANY",
		},
	}
}

func (c BackendApiReqMethod) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BackendApiReqMethod) UnmarshalJSON(b []byte) error {
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

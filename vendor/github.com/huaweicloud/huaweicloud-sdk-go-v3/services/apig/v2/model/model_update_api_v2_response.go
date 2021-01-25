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

// Response Object
type UpdateApiV2Response struct {
	// API名称长度。  为3 ~ 64位的字符串，字符串由中文、英文字母、数字、下划线组成，且只能以英文或中文开头。 > 中文字符必须为UTF-8或者unicode编码。
	Name string `json:"name"`
	// API类型 - 1：公有API - 2：私有API
	Type UpdateApiV2ResponseType `json:"type"`
	// API的版本
	Version *string `json:"version,omitempty"`
	// API的请求协议 - HTTP - HTTPS - BOTH：同时支持HTTP和HTTPS
	ReqProtocol UpdateApiV2ResponseReqProtocol `json:"req_protocol"`
	// API的请求方式
	ReqMethod UpdateApiV2ResponseReqMethod `json:"req_method"`
	// 请求地址。可以包含请求参数，用{}标识，比如/getUserInfo/{userId}，支持 * % - _ . 等特殊字符，总长度不超过512，且满足URI规范。  支持环境变量，使用环境变量时，每个变量名的长度为3 ~ 32位的字符串，字符串由英文字母、数字、中划线、下划线组成，且只能以英文开头。 > 需要服从URI规范。
	ReqUri string `json:"req_uri"`
	// API的认证方式 - NONE：无认证 - APP：APP认证 - IAM：IAM认证 - AUTHORIZER：自定义认证
	AuthType UpdateApiV2ResponseAuthType `json:"auth_type"`
	AuthOpt  *AuthOpt                    `json:"auth_opt,omitempty"`
	// 是否支持跨域 - TRUE：支持 - FALSE：不支持
	Cors *bool `json:"cors,omitempty"`
	// API的匹配方式 - SWA：前缀匹配 - NORMAL：正常匹配（绝对匹配） 默认：NORMAL
	MatchMode *UpdateApiV2ResponseMatchMode `json:"match_mode,omitempty"`
	// 后端类型 - HTTP：web后端 - FUNCTION：函数工作流 - MOCK：模拟的后端
	BackendType UpdateApiV2ResponseBackendType `json:"backend_type"`
	// API描述。字符长度不超过255 > 中文字符必须为UTF-8或者unicode编码。
	Remark *string `json:"remark,omitempty"`
	// API所属的分组编号
	GroupId string `json:"group_id"`
	// API请求体描述，可以是请求体示例、媒体类型、参数等信息。字符长度不超过20480 > 中文字符必须为UTF-8或者unicode编码。
	BodyRemark *string `json:"body_remark,omitempty"`
	// 正常响应示例，描述API的正常返回信息。字符长度不超过20480 > 中文字符必须为UTF-8或者unicode编码。
	ResultNormalSample *string `json:"result_normal_sample,omitempty"`
	// 失败返回示例，描述API的异常返回信息。字符长度不超过20480 > 中文字符必须为UTF-8或者unicode编码。
	ResultFailureSample *string `json:"result_failure_sample,omitempty"`
	// 前端自定义认证对象的ID
	AuthorizerId *string `json:"authorizer_id,omitempty"`
	// 标签。  支持英文，数字，下划线，且只能以英文开头。支持输入多个标签，不同标签以英文逗号分割。
	Tags *[]string `json:"tags,omitempty"`
	// 分组自定义响应ID
	ResponseId *string `json:"response_id,omitempty"`
	// 集成应用ID  暂不支持
	RomaAppId *string `json:"roma_app_id,omitempty"`
	// API绑定的自定义域名  暂不支持
	DomainName *string `json:"domain_name,omitempty"`
	// 标签  待废弃，优先使用tags字段
	Tag *string `json:"tag,omitempty"`
	// API编号
	Id *string `json:"id,omitempty"`
	// API的状态
	Status *int32 `json:"status,omitempty"`
	// 是否需要编排
	ArrangeNecessary *int32 `json:"arrange_necessary,omitempty"`
	// API注册时间
	RegisterTime *sdktime.SdkTime `json:"register_time,omitempty"`
	// API修改时间
	UpdateTime *sdktime.SdkTime `json:"update_time,omitempty"`
	// API所属分组的名称
	GroupName *string `json:"group_name,omitempty"`
	// API所属分组的版本  默认V1，其他版本暂不支持
	GroupVersion *string `json:"group_version,omitempty"`
	// 发布的环境id
	RunEnvId *string `json:"run_env_id,omitempty"`
	// 发布的环境名称
	RunEnvName *string `json:"run_env_name,omitempty"`
	// 发布记录编号  存在多个发布记录时，编号之间用|隔开
	PublishId *string  `json:"publish_id,omitempty"`
	FuncInfo  *ApiFunc `json:"func_info,omitempty"`
	MockInfo  *ApiMock `json:"mock_info,omitempty"`
	// API的请求参数列表
	ReqParams *[]ReqParam `json:"req_params,omitempty"`
	// API的后端参数列表
	BackendParams *[]BackendParam `json:"backend_params,omitempty"`
	// 函数工作流策略后端列表
	PolicyFunctions *[]ApiPolicyFunctionResp `json:"policy_functions,omitempty"`
	// mock策略后端列表
	PolicyMocks *[]ApiPolicyMockResp `json:"policy_mocks,omitempty"`
	BackendApi  *BackendApi          `json:"backend_api,omitempty"`
	// web策略后端列表
	PolicyHttps    *[]ApiPolicyHttpResp `json:"policy_https,omitempty"`
	HttpStatusCode int                  `json:"-"`
}

func (o UpdateApiV2Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateApiV2Response struct{}"
	}

	return strings.Join([]string{"UpdateApiV2Response", string(data)}, " ")
}

type UpdateApiV2ResponseType struct {
	value int32
}

type UpdateApiV2ResponseTypeEnum struct {
	E_1 UpdateApiV2ResponseType
	E_2 UpdateApiV2ResponseType
}

func GetUpdateApiV2ResponseTypeEnum() UpdateApiV2ResponseTypeEnum {
	return UpdateApiV2ResponseTypeEnum{
		E_1: UpdateApiV2ResponseType{
			value: 1,
		}, E_2: UpdateApiV2ResponseType{
			value: 2,
		},
	}
}

func (c UpdateApiV2ResponseType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdateApiV2ResponseType) UnmarshalJSON(b []byte) error {
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

type UpdateApiV2ResponseReqProtocol struct {
	value string
}

type UpdateApiV2ResponseReqProtocolEnum struct {
	HTTP  UpdateApiV2ResponseReqProtocol
	HTTPS UpdateApiV2ResponseReqProtocol
	BOTH  UpdateApiV2ResponseReqProtocol
}

func GetUpdateApiV2ResponseReqProtocolEnum() UpdateApiV2ResponseReqProtocolEnum {
	return UpdateApiV2ResponseReqProtocolEnum{
		HTTP: UpdateApiV2ResponseReqProtocol{
			value: "HTTP",
		},
		HTTPS: UpdateApiV2ResponseReqProtocol{
			value: "HTTPS",
		},
		BOTH: UpdateApiV2ResponseReqProtocol{
			value: "BOTH",
		},
	}
}

func (c UpdateApiV2ResponseReqProtocol) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdateApiV2ResponseReqProtocol) UnmarshalJSON(b []byte) error {
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

type UpdateApiV2ResponseReqMethod struct {
	value string
}

type UpdateApiV2ResponseReqMethodEnum struct {
	GET     UpdateApiV2ResponseReqMethod
	POST    UpdateApiV2ResponseReqMethod
	PUT     UpdateApiV2ResponseReqMethod
	DELETE  UpdateApiV2ResponseReqMethod
	HEAD    UpdateApiV2ResponseReqMethod
	PATCH   UpdateApiV2ResponseReqMethod
	OPTIONS UpdateApiV2ResponseReqMethod
	ANY     UpdateApiV2ResponseReqMethod
}

func GetUpdateApiV2ResponseReqMethodEnum() UpdateApiV2ResponseReqMethodEnum {
	return UpdateApiV2ResponseReqMethodEnum{
		GET: UpdateApiV2ResponseReqMethod{
			value: "GET",
		},
		POST: UpdateApiV2ResponseReqMethod{
			value: "POST",
		},
		PUT: UpdateApiV2ResponseReqMethod{
			value: "PUT",
		},
		DELETE: UpdateApiV2ResponseReqMethod{
			value: "DELETE",
		},
		HEAD: UpdateApiV2ResponseReqMethod{
			value: "HEAD",
		},
		PATCH: UpdateApiV2ResponseReqMethod{
			value: "PATCH",
		},
		OPTIONS: UpdateApiV2ResponseReqMethod{
			value: "OPTIONS",
		},
		ANY: UpdateApiV2ResponseReqMethod{
			value: "ANY",
		},
	}
}

func (c UpdateApiV2ResponseReqMethod) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdateApiV2ResponseReqMethod) UnmarshalJSON(b []byte) error {
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

type UpdateApiV2ResponseAuthType struct {
	value string
}

type UpdateApiV2ResponseAuthTypeEnum struct {
	NONE       UpdateApiV2ResponseAuthType
	APP        UpdateApiV2ResponseAuthType
	IAM        UpdateApiV2ResponseAuthType
	AUTHORIZER UpdateApiV2ResponseAuthType
}

func GetUpdateApiV2ResponseAuthTypeEnum() UpdateApiV2ResponseAuthTypeEnum {
	return UpdateApiV2ResponseAuthTypeEnum{
		NONE: UpdateApiV2ResponseAuthType{
			value: "NONE",
		},
		APP: UpdateApiV2ResponseAuthType{
			value: "APP",
		},
		IAM: UpdateApiV2ResponseAuthType{
			value: "IAM",
		},
		AUTHORIZER: UpdateApiV2ResponseAuthType{
			value: "AUTHORIZER",
		},
	}
}

func (c UpdateApiV2ResponseAuthType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdateApiV2ResponseAuthType) UnmarshalJSON(b []byte) error {
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

type UpdateApiV2ResponseMatchMode struct {
	value string
}

type UpdateApiV2ResponseMatchModeEnum struct {
	SWA    UpdateApiV2ResponseMatchMode
	NORMAL UpdateApiV2ResponseMatchMode
}

func GetUpdateApiV2ResponseMatchModeEnum() UpdateApiV2ResponseMatchModeEnum {
	return UpdateApiV2ResponseMatchModeEnum{
		SWA: UpdateApiV2ResponseMatchMode{
			value: "SWA",
		},
		NORMAL: UpdateApiV2ResponseMatchMode{
			value: "NORMAL",
		},
	}
}

func (c UpdateApiV2ResponseMatchMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdateApiV2ResponseMatchMode) UnmarshalJSON(b []byte) error {
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

type UpdateApiV2ResponseBackendType struct {
	value string
}

type UpdateApiV2ResponseBackendTypeEnum struct {
	HTTP     UpdateApiV2ResponseBackendType
	FUNCTION UpdateApiV2ResponseBackendType
	MOCK     UpdateApiV2ResponseBackendType
}

func GetUpdateApiV2ResponseBackendTypeEnum() UpdateApiV2ResponseBackendTypeEnum {
	return UpdateApiV2ResponseBackendTypeEnum{
		HTTP: UpdateApiV2ResponseBackendType{
			value: "HTTP",
		},
		FUNCTION: UpdateApiV2ResponseBackendType{
			value: "FUNCTION",
		},
		MOCK: UpdateApiV2ResponseBackendType{
			value: "MOCK",
		},
	}
}

func (c UpdateApiV2ResponseBackendType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdateApiV2ResponseBackendType) UnmarshalJSON(b []byte) error {
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

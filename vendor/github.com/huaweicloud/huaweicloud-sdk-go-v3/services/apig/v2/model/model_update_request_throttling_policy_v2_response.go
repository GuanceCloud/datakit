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
type UpdateRequestThrottlingPolicyV2Response struct {
	// 流控绑定的API数量
	BindNum *int32 `json:"bind_num,omitempty"`
	// 是否包含特殊流控配置 - 1：包含 - 2：不包含
	IsIncludeSpecialThrottle *UpdateRequestThrottlingPolicyV2ResponseIsIncludeSpecialThrottle `json:"is_include_special_throttle,omitempty"`
	// 创建时间
	CreateTime *sdktime.SdkTime `json:"create_time,omitempty"`
	// 描述
	Remark *string `json:"remark,omitempty"`
	// 流控策略的类型 - 1：独享，表示绑定到流控策略的单个API流控时间内能够被调用多少次。 - 2：共享，表示绑定到流控策略的所有API流控时间内能够被调用多少次
	Type *UpdateRequestThrottlingPolicyV2ResponseType `json:"type,omitempty"`
	// 流控的时长
	TimeInterval *int32 `json:"time_interval,omitempty"`
	// 单个IP流控时间内能够访问API的次数限制
	IpCallLimits *int32 `json:"ip_call_limits,omitempty"`
	// 单个APP流控时间内能够访问API的次数限制
	AppCallLimits *int32 `json:"app_call_limits,omitempty"`
	// 流控策略的名称
	Name *string `json:"name,omitempty"`
	// 流控的时间单位
	TimeUnit *UpdateRequestThrottlingPolicyV2ResponseTimeUnit `json:"time_unit,omitempty"`
	// 单个API流控时间内能够被访问的次数限制
	ApiCallLimits *int32 `json:"api_call_limits,omitempty"`
	// 流控策略的ID
	Id *string `json:"id,omitempty"`
	// 单个用户流控时间内能够访问API的次数限制
	UserCallLimits *int32 `json:"user_call_limits,omitempty"`
	// 是否开启动态流控  暂不支持
	EnableAdaptiveControl *string `json:"enable_adaptive_control,omitempty"`
	HttpStatusCode        int     `json:"-"`
}

func (o UpdateRequestThrottlingPolicyV2Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateRequestThrottlingPolicyV2Response struct{}"
	}

	return strings.Join([]string{"UpdateRequestThrottlingPolicyV2Response", string(data)}, " ")
}

type UpdateRequestThrottlingPolicyV2ResponseIsIncludeSpecialThrottle struct {
	value int32
}

type UpdateRequestThrottlingPolicyV2ResponseIsIncludeSpecialThrottleEnum struct {
	E_1 UpdateRequestThrottlingPolicyV2ResponseIsIncludeSpecialThrottle
	E_2 UpdateRequestThrottlingPolicyV2ResponseIsIncludeSpecialThrottle
}

func GetUpdateRequestThrottlingPolicyV2ResponseIsIncludeSpecialThrottleEnum() UpdateRequestThrottlingPolicyV2ResponseIsIncludeSpecialThrottleEnum {
	return UpdateRequestThrottlingPolicyV2ResponseIsIncludeSpecialThrottleEnum{
		E_1: UpdateRequestThrottlingPolicyV2ResponseIsIncludeSpecialThrottle{
			value: 1,
		}, E_2: UpdateRequestThrottlingPolicyV2ResponseIsIncludeSpecialThrottle{
			value: 2,
		},
	}
}

func (c UpdateRequestThrottlingPolicyV2ResponseIsIncludeSpecialThrottle) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdateRequestThrottlingPolicyV2ResponseIsIncludeSpecialThrottle) UnmarshalJSON(b []byte) error {
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

type UpdateRequestThrottlingPolicyV2ResponseType struct {
	value int32
}

type UpdateRequestThrottlingPolicyV2ResponseTypeEnum struct {
	E_1 UpdateRequestThrottlingPolicyV2ResponseType
	E_2 UpdateRequestThrottlingPolicyV2ResponseType
}

func GetUpdateRequestThrottlingPolicyV2ResponseTypeEnum() UpdateRequestThrottlingPolicyV2ResponseTypeEnum {
	return UpdateRequestThrottlingPolicyV2ResponseTypeEnum{
		E_1: UpdateRequestThrottlingPolicyV2ResponseType{
			value: 1,
		}, E_2: UpdateRequestThrottlingPolicyV2ResponseType{
			value: 2,
		},
	}
}

func (c UpdateRequestThrottlingPolicyV2ResponseType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdateRequestThrottlingPolicyV2ResponseType) UnmarshalJSON(b []byte) error {
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

type UpdateRequestThrottlingPolicyV2ResponseTimeUnit struct {
	value string
}

type UpdateRequestThrottlingPolicyV2ResponseTimeUnitEnum struct {
	SECOND UpdateRequestThrottlingPolicyV2ResponseTimeUnit
	MINUTE UpdateRequestThrottlingPolicyV2ResponseTimeUnit
	HOUR   UpdateRequestThrottlingPolicyV2ResponseTimeUnit
	DAY    UpdateRequestThrottlingPolicyV2ResponseTimeUnit
}

func GetUpdateRequestThrottlingPolicyV2ResponseTimeUnitEnum() UpdateRequestThrottlingPolicyV2ResponseTimeUnitEnum {
	return UpdateRequestThrottlingPolicyV2ResponseTimeUnitEnum{
		SECOND: UpdateRequestThrottlingPolicyV2ResponseTimeUnit{
			value: "SECOND",
		},
		MINUTE: UpdateRequestThrottlingPolicyV2ResponseTimeUnit{
			value: "MINUTE",
		},
		HOUR: UpdateRequestThrottlingPolicyV2ResponseTimeUnit{
			value: "HOUR",
		},
		DAY: UpdateRequestThrottlingPolicyV2ResponseTimeUnit{
			value: "DAY",
		},
	}
}

func (c UpdateRequestThrottlingPolicyV2ResponseTimeUnit) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdateRequestThrottlingPolicyV2ResponseTimeUnit) UnmarshalJSON(b []byte) error {
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

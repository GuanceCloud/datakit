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
type ShowDetailsOfRequestThrottlingPolicyV2Response struct {
	// 流控绑定的API数量
	BindNum *int32 `json:"bind_num,omitempty"`
	// 是否包含特殊流控配置 - 1：包含 - 2：不包含
	IsIncludeSpecialThrottle *ShowDetailsOfRequestThrottlingPolicyV2ResponseIsIncludeSpecialThrottle `json:"is_include_special_throttle,omitempty"`
	// 创建时间
	CreateTime *sdktime.SdkTime `json:"create_time,omitempty"`
	// 描述
	Remark *string `json:"remark,omitempty"`
	// 流控策略的类型 - 1：独享，表示绑定到流控策略的单个API流控时间内能够被调用多少次。 - 2：共享，表示绑定到流控策略的所有API流控时间内能够被调用多少次
	Type *ShowDetailsOfRequestThrottlingPolicyV2ResponseType `json:"type,omitempty"`
	// 流控的时长
	TimeInterval *int32 `json:"time_interval,omitempty"`
	// 单个IP流控时间内能够访问API的次数限制
	IpCallLimits *int32 `json:"ip_call_limits,omitempty"`
	// 单个APP流控时间内能够访问API的次数限制
	AppCallLimits *int32 `json:"app_call_limits,omitempty"`
	// 流控策略的名称
	Name *string `json:"name,omitempty"`
	// 流控的时间单位
	TimeUnit *ShowDetailsOfRequestThrottlingPolicyV2ResponseTimeUnit `json:"time_unit,omitempty"`
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

func (o ShowDetailsOfRequestThrottlingPolicyV2Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowDetailsOfRequestThrottlingPolicyV2Response struct{}"
	}

	return strings.Join([]string{"ShowDetailsOfRequestThrottlingPolicyV2Response", string(data)}, " ")
}

type ShowDetailsOfRequestThrottlingPolicyV2ResponseIsIncludeSpecialThrottle struct {
	value int32
}

type ShowDetailsOfRequestThrottlingPolicyV2ResponseIsIncludeSpecialThrottleEnum struct {
	E_1 ShowDetailsOfRequestThrottlingPolicyV2ResponseIsIncludeSpecialThrottle
	E_2 ShowDetailsOfRequestThrottlingPolicyV2ResponseIsIncludeSpecialThrottle
}

func GetShowDetailsOfRequestThrottlingPolicyV2ResponseIsIncludeSpecialThrottleEnum() ShowDetailsOfRequestThrottlingPolicyV2ResponseIsIncludeSpecialThrottleEnum {
	return ShowDetailsOfRequestThrottlingPolicyV2ResponseIsIncludeSpecialThrottleEnum{
		E_1: ShowDetailsOfRequestThrottlingPolicyV2ResponseIsIncludeSpecialThrottle{
			value: 1,
		}, E_2: ShowDetailsOfRequestThrottlingPolicyV2ResponseIsIncludeSpecialThrottle{
			value: 2,
		},
	}
}

func (c ShowDetailsOfRequestThrottlingPolicyV2ResponseIsIncludeSpecialThrottle) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowDetailsOfRequestThrottlingPolicyV2ResponseIsIncludeSpecialThrottle) UnmarshalJSON(b []byte) error {
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

type ShowDetailsOfRequestThrottlingPolicyV2ResponseType struct {
	value int32
}

type ShowDetailsOfRequestThrottlingPolicyV2ResponseTypeEnum struct {
	E_1 ShowDetailsOfRequestThrottlingPolicyV2ResponseType
	E_2 ShowDetailsOfRequestThrottlingPolicyV2ResponseType
}

func GetShowDetailsOfRequestThrottlingPolicyV2ResponseTypeEnum() ShowDetailsOfRequestThrottlingPolicyV2ResponseTypeEnum {
	return ShowDetailsOfRequestThrottlingPolicyV2ResponseTypeEnum{
		E_1: ShowDetailsOfRequestThrottlingPolicyV2ResponseType{
			value: 1,
		}, E_2: ShowDetailsOfRequestThrottlingPolicyV2ResponseType{
			value: 2,
		},
	}
}

func (c ShowDetailsOfRequestThrottlingPolicyV2ResponseType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowDetailsOfRequestThrottlingPolicyV2ResponseType) UnmarshalJSON(b []byte) error {
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

type ShowDetailsOfRequestThrottlingPolicyV2ResponseTimeUnit struct {
	value string
}

type ShowDetailsOfRequestThrottlingPolicyV2ResponseTimeUnitEnum struct {
	SECOND ShowDetailsOfRequestThrottlingPolicyV2ResponseTimeUnit
	MINUTE ShowDetailsOfRequestThrottlingPolicyV2ResponseTimeUnit
	HOUR   ShowDetailsOfRequestThrottlingPolicyV2ResponseTimeUnit
	DAY    ShowDetailsOfRequestThrottlingPolicyV2ResponseTimeUnit
}

func GetShowDetailsOfRequestThrottlingPolicyV2ResponseTimeUnitEnum() ShowDetailsOfRequestThrottlingPolicyV2ResponseTimeUnitEnum {
	return ShowDetailsOfRequestThrottlingPolicyV2ResponseTimeUnitEnum{
		SECOND: ShowDetailsOfRequestThrottlingPolicyV2ResponseTimeUnit{
			value: "SECOND",
		},
		MINUTE: ShowDetailsOfRequestThrottlingPolicyV2ResponseTimeUnit{
			value: "MINUTE",
		},
		HOUR: ShowDetailsOfRequestThrottlingPolicyV2ResponseTimeUnit{
			value: "HOUR",
		},
		DAY: ShowDetailsOfRequestThrottlingPolicyV2ResponseTimeUnit{
			value: "DAY",
		},
	}
}

func (c ShowDetailsOfRequestThrottlingPolicyV2ResponseTimeUnit) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowDetailsOfRequestThrottlingPolicyV2ResponseTimeUnit) UnmarshalJSON(b []byte) error {
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

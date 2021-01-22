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

type ThrottleResp struct {
	// 流控绑定的API数量
	BindNum *int32 `json:"bind_num,omitempty"`
	// 是否包含特殊流控配置 - 1：包含 - 2：不包含
	IsIncludeSpecialThrottle *ThrottleRespIsIncludeSpecialThrottle `json:"is_include_special_throttle,omitempty"`
	// 创建时间
	CreateTime *sdktime.SdkTime `json:"create_time,omitempty"`
	// 描述
	Remark *string `json:"remark,omitempty"`
	// 流控策略的类型 - 1：独享，表示绑定到流控策略的单个API流控时间内能够被调用多少次。 - 2：共享，表示绑定到流控策略的所有API流控时间内能够被调用多少次
	Type *ThrottleRespType `json:"type,omitempty"`
	// 流控的时长
	TimeInterval *int32 `json:"time_interval,omitempty"`
	// 单个IP流控时间内能够访问API的次数限制
	IpCallLimits *int32 `json:"ip_call_limits,omitempty"`
	// 单个APP流控时间内能够访问API的次数限制
	AppCallLimits *int32 `json:"app_call_limits,omitempty"`
	// 流控策略的名称
	Name *string `json:"name,omitempty"`
	// 流控的时间单位
	TimeUnit *ThrottleRespTimeUnit `json:"time_unit,omitempty"`
	// 单个API流控时间内能够被访问的次数限制
	ApiCallLimits *int32 `json:"api_call_limits,omitempty"`
	// 流控策略的ID
	Id *string `json:"id,omitempty"`
	// 单个用户流控时间内能够访问API的次数限制
	UserCallLimits *int32 `json:"user_call_limits,omitempty"`
	// 是否开启动态流控  暂不支持
	EnableAdaptiveControl *string `json:"enable_adaptive_control,omitempty"`
}

func (o ThrottleResp) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ThrottleResp struct{}"
	}

	return strings.Join([]string{"ThrottleResp", string(data)}, " ")
}

type ThrottleRespIsIncludeSpecialThrottle struct {
	value int32
}

type ThrottleRespIsIncludeSpecialThrottleEnum struct {
	E_1 ThrottleRespIsIncludeSpecialThrottle
	E_2 ThrottleRespIsIncludeSpecialThrottle
}

func GetThrottleRespIsIncludeSpecialThrottleEnum() ThrottleRespIsIncludeSpecialThrottleEnum {
	return ThrottleRespIsIncludeSpecialThrottleEnum{
		E_1: ThrottleRespIsIncludeSpecialThrottle{
			value: 1,
		}, E_2: ThrottleRespIsIncludeSpecialThrottle{
			value: 2,
		},
	}
}

func (c ThrottleRespIsIncludeSpecialThrottle) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ThrottleRespIsIncludeSpecialThrottle) UnmarshalJSON(b []byte) error {
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

type ThrottleRespType struct {
	value int32
}

type ThrottleRespTypeEnum struct {
	E_1 ThrottleRespType
	E_2 ThrottleRespType
}

func GetThrottleRespTypeEnum() ThrottleRespTypeEnum {
	return ThrottleRespTypeEnum{
		E_1: ThrottleRespType{
			value: 1,
		}, E_2: ThrottleRespType{
			value: 2,
		},
	}
}

func (c ThrottleRespType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ThrottleRespType) UnmarshalJSON(b []byte) error {
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

type ThrottleRespTimeUnit struct {
	value string
}

type ThrottleRespTimeUnitEnum struct {
	SECOND ThrottleRespTimeUnit
	MINUTE ThrottleRespTimeUnit
	HOUR   ThrottleRespTimeUnit
	DAY    ThrottleRespTimeUnit
}

func GetThrottleRespTimeUnitEnum() ThrottleRespTimeUnitEnum {
	return ThrottleRespTimeUnitEnum{
		SECOND: ThrottleRespTimeUnit{
			value: "SECOND",
		},
		MINUTE: ThrottleRespTimeUnit{
			value: "MINUTE",
		},
		HOUR: ThrottleRespTimeUnit{
			value: "HOUR",
		},
		DAY: ThrottleRespTimeUnit{
			value: "DAY",
		},
	}
}

func (c ThrottleRespTimeUnit) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ThrottleRespTimeUnit) UnmarshalJSON(b []byte) error {
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

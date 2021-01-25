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

type ThrottleReq struct {
	// APP流量限制是指一个API在时长之内被每个APP访问的次数上限，该数值不超过用户流量限制值。输入的值不超过2147483647。正整数。
	AppCallLimits *int32 `json:"app_call_limits,omitempty"`
	// 流控策略名称。支持汉字，英文，数字，下划线，且只能以英文和汉字开头，3 ~ 64字符。 > 中文字符必须为UTF-8或者unicode编码。
	Name string `json:"name"`
	// 流控的时间单位
	TimeUnit ThrottleReqTimeUnit `json:"time_unit"`
	// 流控策略描述字符长度不超过255。 > 中文字符必须为UTF-8或者unicode编码。
	Remark *string `json:"remark,omitempty"`
	// API流量限制是指时长内一个API能够被访问的次数上限。该值不超过系统默认配额限制，系统默认配额为200tps，用户可根据实际情况修改该系统默认配额。输入的值不超过2147483647。正整数。
	ApiCallLimits int32 `json:"api_call_limits"`
	// 流控策略的类型 - 1：独享，表示绑定到流控策略的单个API流控时间内能够被调用多少次。 - 2：共享，表示绑定到流控策略的所有API流控时间内能够被调用多少次。
	Type *ThrottleReqType `json:"type,omitempty"`
	// 是否开启动态流控： - TRUE - FALSE  暂不支持
	EnableAdaptiveControl *string `json:"enable_adaptive_control,omitempty"`
	// 用户流量限制是指一个API在时长之内每一个用户能访问的次数上限，该数值不超过API流量限制值。输入的值不超过2147483647。正整数。
	UserCallLimits *int32 `json:"user_call_limits,omitempty"`
	// 流量控制的时长单位。与“流量限制次数”配合使用，表示单位时间内的API请求次数上限。输入的值不超过2147483647。正整数。
	TimeInterval int32 `json:"time_interval"`
	// 源IP流量限制是指一个API在时长之内被每个IP访问的次数上限，该数值不超过API流量限制值。输入的值不超过2147483647。正整数。
	IpCallLimits *int32 `json:"ip_call_limits,omitempty"`
}

func (o ThrottleReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ThrottleReq struct{}"
	}

	return strings.Join([]string{"ThrottleReq", string(data)}, " ")
}

type ThrottleReqTimeUnit struct {
	value string
}

type ThrottleReqTimeUnitEnum struct {
	SECOND ThrottleReqTimeUnit
	MINUTE ThrottleReqTimeUnit
	HOUR   ThrottleReqTimeUnit
	DAY    ThrottleReqTimeUnit
}

func GetThrottleReqTimeUnitEnum() ThrottleReqTimeUnitEnum {
	return ThrottleReqTimeUnitEnum{
		SECOND: ThrottleReqTimeUnit{
			value: "SECOND",
		},
		MINUTE: ThrottleReqTimeUnit{
			value: "MINUTE",
		},
		HOUR: ThrottleReqTimeUnit{
			value: "HOUR",
		},
		DAY: ThrottleReqTimeUnit{
			value: "DAY",
		},
	}
}

func (c ThrottleReqTimeUnit) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ThrottleReqTimeUnit) UnmarshalJSON(b []byte) error {
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

type ThrottleReqType struct {
	value int32
}

type ThrottleReqTypeEnum struct {
	E_1 ThrottleReqType
	E_2 ThrottleReqType
}

func GetThrottleReqTypeEnum() ThrottleReqTypeEnum {
	return ThrottleReqTypeEnum{
		E_1: ThrottleReqType{
			value: 1,
		}, E_2: ThrottleReqType{
			value: 2,
		},
	}
}

func (c ThrottleReqType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ThrottleReqType) UnmarshalJSON(b []byte) error {
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

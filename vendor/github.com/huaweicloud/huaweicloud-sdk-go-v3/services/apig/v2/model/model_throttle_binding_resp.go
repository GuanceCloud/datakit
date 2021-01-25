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

type ThrottleBindingResp struct {
	// API的发布记录编号
	PublishId *string `json:"publish_id,omitempty"`
	// 策略作用域，取值如下： - 1：整个API - 2： 单个用户 - 3：单个APP  目前只支持1
	Scope *ThrottleBindingRespScope `json:"scope,omitempty"`
	// 流控策略的ID
	StrategyId *string `json:"strategy_id,omitempty"`
	// 绑定时间
	ApplyTime *sdktime.SdkTime `json:"apply_time,omitempty"`
	// 绑定关系的ID
	Id *string `json:"id,omitempty"`
}

func (o ThrottleBindingResp) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ThrottleBindingResp struct{}"
	}

	return strings.Join([]string{"ThrottleBindingResp", string(data)}, " ")
}

type ThrottleBindingRespScope struct {
	value int32
}

type ThrottleBindingRespScopeEnum struct {
	E_1 ThrottleBindingRespScope
	E_2 ThrottleBindingRespScope
	E_3 ThrottleBindingRespScope
}

func GetThrottleBindingRespScopeEnum() ThrottleBindingRespScopeEnum {
	return ThrottleBindingRespScopeEnum{
		E_1: ThrottleBindingRespScope{
			value: 1,
		}, E_2: ThrottleBindingRespScope{
			value: 2,
		}, E_3: ThrottleBindingRespScope{
			value: 3,
		},
	}
}

func (c ThrottleBindingRespScope) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ThrottleBindingRespScope) UnmarshalJSON(b []byte) error {
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

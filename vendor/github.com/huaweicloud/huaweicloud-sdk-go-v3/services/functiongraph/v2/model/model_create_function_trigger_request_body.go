/*
 * FunctionGraph
 *
 * API v2
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

type CreateFunctionTriggerRequestBody struct {
	// 触发器类型。  - TIMER: 定时触发器。 - APIG: APIGW触发器。 - CTS: 云审计触发器，需要先开通云审计服务。 - DDS: 文档数据库触发器，需要开启函数vpc。 - DMS: 分布式消息服务触发器，需要配置dms委托。 - DIS: 数据接入服务触发器，需要配置dis委托。 - LTS: 云审计日志服务触发器，需要配置lts委托。 - OBS: 对象存储服务触发器。 - KAFKA: 专享版本kafka触发器。
	TriggerTypeCode CreateFunctionTriggerRequestBodyTriggerTypeCode `json:"trigger_type_code"`
	// 触发器状态，取值为ACTIVE,DISABLED。
	TriggerStatus *CreateFunctionTriggerRequestBodyTriggerStatus `json:"trigger_status,omitempty"`
	// 消息代码。
	EventTypeCode *string `json:"event_type_code,omitempty"`
	// 事件结构体。
	EventData *interface{} `json:"event_data"`
}

func (o CreateFunctionTriggerRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateFunctionTriggerRequestBody struct{}"
	}

	return strings.Join([]string{"CreateFunctionTriggerRequestBody", string(data)}, " ")
}

type CreateFunctionTriggerRequestBodyTriggerTypeCode struct {
	value string
}

type CreateFunctionTriggerRequestBodyTriggerTypeCodeEnum struct {
	TIMER CreateFunctionTriggerRequestBodyTriggerTypeCode
	APIG  CreateFunctionTriggerRequestBodyTriggerTypeCode
	CTS   CreateFunctionTriggerRequestBodyTriggerTypeCode
	DDS   CreateFunctionTriggerRequestBodyTriggerTypeCode
	DMS   CreateFunctionTriggerRequestBodyTriggerTypeCode
	DIS   CreateFunctionTriggerRequestBodyTriggerTypeCode
	LTS   CreateFunctionTriggerRequestBodyTriggerTypeCode
	OBS   CreateFunctionTriggerRequestBodyTriggerTypeCode
	KAFKA CreateFunctionTriggerRequestBodyTriggerTypeCode
}

func GetCreateFunctionTriggerRequestBodyTriggerTypeCodeEnum() CreateFunctionTriggerRequestBodyTriggerTypeCodeEnum {
	return CreateFunctionTriggerRequestBodyTriggerTypeCodeEnum{
		TIMER: CreateFunctionTriggerRequestBodyTriggerTypeCode{
			value: "TIMER",
		},
		APIG: CreateFunctionTriggerRequestBodyTriggerTypeCode{
			value: "APIG",
		},
		CTS: CreateFunctionTriggerRequestBodyTriggerTypeCode{
			value: "CTS",
		},
		DDS: CreateFunctionTriggerRequestBodyTriggerTypeCode{
			value: "DDS",
		},
		DMS: CreateFunctionTriggerRequestBodyTriggerTypeCode{
			value: "DMS",
		},
		DIS: CreateFunctionTriggerRequestBodyTriggerTypeCode{
			value: "DIS",
		},
		LTS: CreateFunctionTriggerRequestBodyTriggerTypeCode{
			value: "LTS",
		},
		OBS: CreateFunctionTriggerRequestBodyTriggerTypeCode{
			value: "OBS",
		},
		KAFKA: CreateFunctionTriggerRequestBodyTriggerTypeCode{
			value: "KAFKA",
		},
	}
}

func (c CreateFunctionTriggerRequestBodyTriggerTypeCode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateFunctionTriggerRequestBodyTriggerTypeCode) UnmarshalJSON(b []byte) error {
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

type CreateFunctionTriggerRequestBodyTriggerStatus struct {
	value string
}

type CreateFunctionTriggerRequestBodyTriggerStatusEnum struct {
	ACTIVE   CreateFunctionTriggerRequestBodyTriggerStatus
	DISABLED CreateFunctionTriggerRequestBodyTriggerStatus
}

func GetCreateFunctionTriggerRequestBodyTriggerStatusEnum() CreateFunctionTriggerRequestBodyTriggerStatusEnum {
	return CreateFunctionTriggerRequestBodyTriggerStatusEnum{
		ACTIVE: CreateFunctionTriggerRequestBodyTriggerStatus{
			value: "ACTIVE",
		},
		DISABLED: CreateFunctionTriggerRequestBodyTriggerStatus{
			value: "DISABLED",
		},
	}
}

func (c CreateFunctionTriggerRequestBodyTriggerStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateFunctionTriggerRequestBodyTriggerStatus) UnmarshalJSON(b []byte) error {
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

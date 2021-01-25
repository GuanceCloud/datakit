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
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/sdktime"
	"strings"
)

// Response Object
type ShowFunctionTriggerResponse struct {
	// 触发器ID。
	TriggerId *string `json:"trigger_id,omitempty"`
	// 触发器类型。  - TIMER: \"定时触发器。\" - APIG: \"APIG触发器。\" - CTS: \"云审计服务触发器。\" - DDS: \"文档数据库服务触发器。\" - DMS: \"分布式服务触发器。\" - DIS: \"数据接入服务触发器。\" - LTS: \"云日志服务触发器。\" - OBS: \"对象存储触发器。\" - SMN: \"消息通知服务触发器。\" - KAFKA: \"专享版消息通知服务触发器。\"
	TriggerTypeCode *ShowFunctionTriggerResponseTriggerTypeCode `json:"trigger_type_code,omitempty"`
	// \"触发器状态\"  - ACTIVE: 启用状态。 - DISABLED: 禁用状态。
	TriggerStatus *ShowFunctionTriggerResponseTriggerStatus `json:"trigger_status,omitempty"`
	// 触发器源事件。
	EventData *interface{} `json:"event_data,omitempty"`
	// 最后更新时间。
	LastUpdatedTime *sdktime.SdkTime `json:"last_updated_time,omitempty"`
	// 触发器创建时间。
	CreatedTime    *sdktime.SdkTime `json:"created_time,omitempty"`
	HttpStatusCode int              `json:"-"`
}

func (o ShowFunctionTriggerResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowFunctionTriggerResponse struct{}"
	}

	return strings.Join([]string{"ShowFunctionTriggerResponse", string(data)}, " ")
}

type ShowFunctionTriggerResponseTriggerTypeCode struct {
	value string
}

type ShowFunctionTriggerResponseTriggerTypeCodeEnum struct {
	TIMER ShowFunctionTriggerResponseTriggerTypeCode
	APIG  ShowFunctionTriggerResponseTriggerTypeCode
	CTS   ShowFunctionTriggerResponseTriggerTypeCode
	DDS   ShowFunctionTriggerResponseTriggerTypeCode
	DMS   ShowFunctionTriggerResponseTriggerTypeCode
	DIS   ShowFunctionTriggerResponseTriggerTypeCode
	LTS   ShowFunctionTriggerResponseTriggerTypeCode
	OBS   ShowFunctionTriggerResponseTriggerTypeCode
	SMN   ShowFunctionTriggerResponseTriggerTypeCode
	KAFKA ShowFunctionTriggerResponseTriggerTypeCode
}

func GetShowFunctionTriggerResponseTriggerTypeCodeEnum() ShowFunctionTriggerResponseTriggerTypeCodeEnum {
	return ShowFunctionTriggerResponseTriggerTypeCodeEnum{
		TIMER: ShowFunctionTriggerResponseTriggerTypeCode{
			value: "TIMER",
		},
		APIG: ShowFunctionTriggerResponseTriggerTypeCode{
			value: "APIG",
		},
		CTS: ShowFunctionTriggerResponseTriggerTypeCode{
			value: "CTS",
		},
		DDS: ShowFunctionTriggerResponseTriggerTypeCode{
			value: "DDS",
		},
		DMS: ShowFunctionTriggerResponseTriggerTypeCode{
			value: "DMS",
		},
		DIS: ShowFunctionTriggerResponseTriggerTypeCode{
			value: "DIS",
		},
		LTS: ShowFunctionTriggerResponseTriggerTypeCode{
			value: "LTS",
		},
		OBS: ShowFunctionTriggerResponseTriggerTypeCode{
			value: "OBS",
		},
		SMN: ShowFunctionTriggerResponseTriggerTypeCode{
			value: "SMN",
		},
		KAFKA: ShowFunctionTriggerResponseTriggerTypeCode{
			value: "KAFKA",
		},
	}
}

func (c ShowFunctionTriggerResponseTriggerTypeCode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowFunctionTriggerResponseTriggerTypeCode) UnmarshalJSON(b []byte) error {
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

type ShowFunctionTriggerResponseTriggerStatus struct {
	value string
}

type ShowFunctionTriggerResponseTriggerStatusEnum struct {
	ACTIVE   ShowFunctionTriggerResponseTriggerStatus
	DISABLED ShowFunctionTriggerResponseTriggerStatus
}

func GetShowFunctionTriggerResponseTriggerStatusEnum() ShowFunctionTriggerResponseTriggerStatusEnum {
	return ShowFunctionTriggerResponseTriggerStatusEnum{
		ACTIVE: ShowFunctionTriggerResponseTriggerStatus{
			value: "ACTIVE",
		},
		DISABLED: ShowFunctionTriggerResponseTriggerStatus{
			value: "DISABLED",
		},
	}
}

func (c ShowFunctionTriggerResponseTriggerStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowFunctionTriggerResponseTriggerStatus) UnmarshalJSON(b []byte) error {
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

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

type UpdateTriggerRequestBody struct {
	// 触发器状态
	TriggerStatus *UpdateTriggerRequestBodyTriggerStatus `json:"trigger_status,omitempty"`
}

func (o UpdateTriggerRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateTriggerRequestBody struct{}"
	}

	return strings.Join([]string{"UpdateTriggerRequestBody", string(data)}, " ")
}

type UpdateTriggerRequestBodyTriggerStatus struct {
	value string
}

type UpdateTriggerRequestBodyTriggerStatusEnum struct {
	ACTIVE   UpdateTriggerRequestBodyTriggerStatus
	DISABLED UpdateTriggerRequestBodyTriggerStatus
}

func GetUpdateTriggerRequestBodyTriggerStatusEnum() UpdateTriggerRequestBodyTriggerStatusEnum {
	return UpdateTriggerRequestBodyTriggerStatusEnum{
		ACTIVE: UpdateTriggerRequestBodyTriggerStatus{
			value: "ACTIVE",
		},
		DISABLED: UpdateTriggerRequestBodyTriggerStatus{
			value: "DISABLED",
		},
	}
}

func (c UpdateTriggerRequestBodyTriggerStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdateTriggerRequestBodyTriggerStatus) UnmarshalJSON(b []byte) error {
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

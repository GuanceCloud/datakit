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

// Request Object
type UpdateTriggerRequest struct {
	FunctionUrn     string                              `json:"function_urn"`
	TriggerTypeCode UpdateTriggerRequestTriggerTypeCode `json:"trigger_type_code"`
	TriggerId       string                              `json:"triggerId"`
	Body            *UpdateTriggerRequestBody           `json:"body,omitempty"`
}

func (o UpdateTriggerRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateTriggerRequest struct{}"
	}

	return strings.Join([]string{"UpdateTriggerRequest", string(data)}, " ")
}

type UpdateTriggerRequestTriggerTypeCode struct {
	value string
}

type UpdateTriggerRequestTriggerTypeCodeEnum struct {
	TIMER UpdateTriggerRequestTriggerTypeCode
	APIG  UpdateTriggerRequestTriggerTypeCode
	CTS   UpdateTriggerRequestTriggerTypeCode
	DDS   UpdateTriggerRequestTriggerTypeCode
	DMS   UpdateTriggerRequestTriggerTypeCode
	DIS   UpdateTriggerRequestTriggerTypeCode
	LTS   UpdateTriggerRequestTriggerTypeCode
	OBS   UpdateTriggerRequestTriggerTypeCode
	SMN   UpdateTriggerRequestTriggerTypeCode
	KAFKA UpdateTriggerRequestTriggerTypeCode
}

func GetUpdateTriggerRequestTriggerTypeCodeEnum() UpdateTriggerRequestTriggerTypeCodeEnum {
	return UpdateTriggerRequestTriggerTypeCodeEnum{
		TIMER: UpdateTriggerRequestTriggerTypeCode{
			value: "TIMER",
		},
		APIG: UpdateTriggerRequestTriggerTypeCode{
			value: "APIG",
		},
		CTS: UpdateTriggerRequestTriggerTypeCode{
			value: "CTS",
		},
		DDS: UpdateTriggerRequestTriggerTypeCode{
			value: "DDS",
		},
		DMS: UpdateTriggerRequestTriggerTypeCode{
			value: "DMS",
		},
		DIS: UpdateTriggerRequestTriggerTypeCode{
			value: "DIS",
		},
		LTS: UpdateTriggerRequestTriggerTypeCode{
			value: "LTS",
		},
		OBS: UpdateTriggerRequestTriggerTypeCode{
			value: "OBS",
		},
		SMN: UpdateTriggerRequestTriggerTypeCode{
			value: "SMN",
		},
		KAFKA: UpdateTriggerRequestTriggerTypeCode{
			value: "KAFKA",
		},
	}
}

func (c UpdateTriggerRequestTriggerTypeCode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *UpdateTriggerRequestTriggerTypeCode) UnmarshalJSON(b []byte) error {
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

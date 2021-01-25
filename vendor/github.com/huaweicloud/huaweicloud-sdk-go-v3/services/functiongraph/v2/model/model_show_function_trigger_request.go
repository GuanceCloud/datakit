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
type ShowFunctionTriggerRequest struct {
	FunctionUrn     string                                    `json:"function_urn"`
	TriggerTypeCode ShowFunctionTriggerRequestTriggerTypeCode `json:"trigger_type_code"`
	TriggerId       string                                    `json:"triggerId"`
}

func (o ShowFunctionTriggerRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowFunctionTriggerRequest struct{}"
	}

	return strings.Join([]string{"ShowFunctionTriggerRequest", string(data)}, " ")
}

type ShowFunctionTriggerRequestTriggerTypeCode struct {
	value string
}

type ShowFunctionTriggerRequestTriggerTypeCodeEnum struct {
	TIMER ShowFunctionTriggerRequestTriggerTypeCode
	APIG  ShowFunctionTriggerRequestTriggerTypeCode
	CTS   ShowFunctionTriggerRequestTriggerTypeCode
	DDS   ShowFunctionTriggerRequestTriggerTypeCode
	DMS   ShowFunctionTriggerRequestTriggerTypeCode
	DIS   ShowFunctionTriggerRequestTriggerTypeCode
	LTS   ShowFunctionTriggerRequestTriggerTypeCode
	OBS   ShowFunctionTriggerRequestTriggerTypeCode
	SMN   ShowFunctionTriggerRequestTriggerTypeCode
	KAFKA ShowFunctionTriggerRequestTriggerTypeCode
}

func GetShowFunctionTriggerRequestTriggerTypeCodeEnum() ShowFunctionTriggerRequestTriggerTypeCodeEnum {
	return ShowFunctionTriggerRequestTriggerTypeCodeEnum{
		TIMER: ShowFunctionTriggerRequestTriggerTypeCode{
			value: "TIMER",
		},
		APIG: ShowFunctionTriggerRequestTriggerTypeCode{
			value: "APIG",
		},
		CTS: ShowFunctionTriggerRequestTriggerTypeCode{
			value: "CTS",
		},
		DDS: ShowFunctionTriggerRequestTriggerTypeCode{
			value: "DDS",
		},
		DMS: ShowFunctionTriggerRequestTriggerTypeCode{
			value: "DMS",
		},
		DIS: ShowFunctionTriggerRequestTriggerTypeCode{
			value: "DIS",
		},
		LTS: ShowFunctionTriggerRequestTriggerTypeCode{
			value: "LTS",
		},
		OBS: ShowFunctionTriggerRequestTriggerTypeCode{
			value: "OBS",
		},
		SMN: ShowFunctionTriggerRequestTriggerTypeCode{
			value: "SMN",
		},
		KAFKA: ShowFunctionTriggerRequestTriggerTypeCode{
			value: "KAFKA",
		},
	}
}

func (c ShowFunctionTriggerRequestTriggerTypeCode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowFunctionTriggerRequestTriggerTypeCode) UnmarshalJSON(b []byte) error {
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

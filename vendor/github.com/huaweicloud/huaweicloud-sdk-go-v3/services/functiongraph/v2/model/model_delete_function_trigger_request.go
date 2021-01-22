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
type DeleteFunctionTriggerRequest struct {
	FunctionUrn     string                                      `json:"function_urn"`
	TriggerTypeCode DeleteFunctionTriggerRequestTriggerTypeCode `json:"trigger_type_code"`
	TriggerId       string                                      `json:"triggerId"`
}

func (o DeleteFunctionTriggerRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteFunctionTriggerRequest struct{}"
	}

	return strings.Join([]string{"DeleteFunctionTriggerRequest", string(data)}, " ")
}

type DeleteFunctionTriggerRequestTriggerTypeCode struct {
	value string
}

type DeleteFunctionTriggerRequestTriggerTypeCodeEnum struct {
	TIMER DeleteFunctionTriggerRequestTriggerTypeCode
	APIG  DeleteFunctionTriggerRequestTriggerTypeCode
	CTS   DeleteFunctionTriggerRequestTriggerTypeCode
	DDS   DeleteFunctionTriggerRequestTriggerTypeCode
	DMS   DeleteFunctionTriggerRequestTriggerTypeCode
	DIS   DeleteFunctionTriggerRequestTriggerTypeCode
	LTS   DeleteFunctionTriggerRequestTriggerTypeCode
	OBS   DeleteFunctionTriggerRequestTriggerTypeCode
	SMN   DeleteFunctionTriggerRequestTriggerTypeCode
	KAFKA DeleteFunctionTriggerRequestTriggerTypeCode
}

func GetDeleteFunctionTriggerRequestTriggerTypeCodeEnum() DeleteFunctionTriggerRequestTriggerTypeCodeEnum {
	return DeleteFunctionTriggerRequestTriggerTypeCodeEnum{
		TIMER: DeleteFunctionTriggerRequestTriggerTypeCode{
			value: "TIMER",
		},
		APIG: DeleteFunctionTriggerRequestTriggerTypeCode{
			value: "APIG",
		},
		CTS: DeleteFunctionTriggerRequestTriggerTypeCode{
			value: "CTS",
		},
		DDS: DeleteFunctionTriggerRequestTriggerTypeCode{
			value: "DDS",
		},
		DMS: DeleteFunctionTriggerRequestTriggerTypeCode{
			value: "DMS",
		},
		DIS: DeleteFunctionTriggerRequestTriggerTypeCode{
			value: "DIS",
		},
		LTS: DeleteFunctionTriggerRequestTriggerTypeCode{
			value: "LTS",
		},
		OBS: DeleteFunctionTriggerRequestTriggerTypeCode{
			value: "OBS",
		},
		SMN: DeleteFunctionTriggerRequestTriggerTypeCode{
			value: "SMN",
		},
		KAFKA: DeleteFunctionTriggerRequestTriggerTypeCode{
			value: "KAFKA",
		},
	}
}

func (c DeleteFunctionTriggerRequestTriggerTypeCode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DeleteFunctionTriggerRequestTriggerTypeCode) UnmarshalJSON(b []byte) error {
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

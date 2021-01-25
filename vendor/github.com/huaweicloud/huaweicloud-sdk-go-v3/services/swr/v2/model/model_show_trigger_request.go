/*
 * SWR
 *
 * SWR API
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
type ShowTriggerRequest struct {
	ContentType ShowTriggerRequestContentType `json:"Content-Type"`
	Namespace   string                        `json:"namespace"`
	Repository  string                        `json:"repository"`
	Trigger     string                        `json:"trigger"`
}

func (o ShowTriggerRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowTriggerRequest struct{}"
	}

	return strings.Join([]string{"ShowTriggerRequest", string(data)}, " ")
}

type ShowTriggerRequestContentType struct {
	value string
}

type ShowTriggerRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 ShowTriggerRequestContentType
	APPLICATION_JSON             ShowTriggerRequestContentType
}

func GetShowTriggerRequestContentTypeEnum() ShowTriggerRequestContentTypeEnum {
	return ShowTriggerRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: ShowTriggerRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: ShowTriggerRequestContentType{
			value: "application/json",
		},
	}
}

func (c ShowTriggerRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowTriggerRequestContentType) UnmarshalJSON(b []byte) error {
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

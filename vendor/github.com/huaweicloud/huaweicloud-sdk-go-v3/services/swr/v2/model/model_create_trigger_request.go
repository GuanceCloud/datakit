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
type CreateTriggerRequest struct {
	ContentType CreateTriggerRequestContentType `json:"Content-Type"`
	Namespace   string                          `json:"namespace"`
	Repository  string                          `json:"repository"`
	Body        *CreateTriggerRequestBody       `json:"body,omitempty"`
}

func (o CreateTriggerRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateTriggerRequest struct{}"
	}

	return strings.Join([]string{"CreateTriggerRequest", string(data)}, " ")
}

type CreateTriggerRequestContentType struct {
	value string
}

type CreateTriggerRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 CreateTriggerRequestContentType
	APPLICATION_JSON             CreateTriggerRequestContentType
}

func GetCreateTriggerRequestContentTypeEnum() CreateTriggerRequestContentTypeEnum {
	return CreateTriggerRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: CreateTriggerRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: CreateTriggerRequestContentType{
			value: "application/json",
		},
	}
}

func (c CreateTriggerRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateTriggerRequestContentType) UnmarshalJSON(b []byte) error {
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

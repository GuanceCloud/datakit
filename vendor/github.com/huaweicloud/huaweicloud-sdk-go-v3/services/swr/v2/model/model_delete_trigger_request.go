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
type DeleteTriggerRequest struct {
	ContentType DeleteTriggerRequestContentType `json:"Content-Type"`
	Namespace   string                          `json:"namespace"`
	Repository  string                          `json:"repository"`
	Trigger     string                          `json:"trigger"`
}

func (o DeleteTriggerRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteTriggerRequest struct{}"
	}

	return strings.Join([]string{"DeleteTriggerRequest", string(data)}, " ")
}

type DeleteTriggerRequestContentType struct {
	value string
}

type DeleteTriggerRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 DeleteTriggerRequestContentType
	APPLICATION_JSON             DeleteTriggerRequestContentType
}

func GetDeleteTriggerRequestContentTypeEnum() DeleteTriggerRequestContentTypeEnum {
	return DeleteTriggerRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: DeleteTriggerRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: DeleteTriggerRequestContentType{
			value: "application/json",
		},
	}
}

func (c DeleteTriggerRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *DeleteTriggerRequestContentType) UnmarshalJSON(b []byte) error {
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

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
type ListTriggersDetailsRequest struct {
	ContentType ListTriggersDetailsRequestContentType `json:"Content-Type"`
	Namespace   string                                `json:"namespace"`
	Repository  string                                `json:"repository"`
}

func (o ListTriggersDetailsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListTriggersDetailsRequest struct{}"
	}

	return strings.Join([]string{"ListTriggersDetailsRequest", string(data)}, " ")
}

type ListTriggersDetailsRequestContentType struct {
	value string
}

type ListTriggersDetailsRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 ListTriggersDetailsRequestContentType
	APPLICATION_JSON             ListTriggersDetailsRequestContentType
}

func GetListTriggersDetailsRequestContentTypeEnum() ListTriggersDetailsRequestContentTypeEnum {
	return ListTriggersDetailsRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: ListTriggersDetailsRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: ListTriggersDetailsRequestContentType{
			value: "application/json",
		},
	}
}

func (c ListTriggersDetailsRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListTriggersDetailsRequestContentType) UnmarshalJSON(b []byte) error {
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

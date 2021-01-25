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
type ListRetentionsRequest struct {
	ContentType ListRetentionsRequestContentType `json:"Content-Type"`
	Namespace   string                           `json:"namespace"`
	Repository  string                           `json:"repository"`
}

func (o ListRetentionsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListRetentionsRequest struct{}"
	}

	return strings.Join([]string{"ListRetentionsRequest", string(data)}, " ")
}

type ListRetentionsRequestContentType struct {
	value string
}

type ListRetentionsRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 ListRetentionsRequestContentType
	APPLICATION_JSON             ListRetentionsRequestContentType
}

func GetListRetentionsRequestContentTypeEnum() ListRetentionsRequestContentTypeEnum {
	return ListRetentionsRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: ListRetentionsRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: ListRetentionsRequestContentType{
			value: "application/json",
		},
	}
}

func (c ListRetentionsRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListRetentionsRequestContentType) UnmarshalJSON(b []byte) error {
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

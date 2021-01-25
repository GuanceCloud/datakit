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
type ListRetentionHistoriesRequest struct {
	ContentType ListRetentionHistoriesRequestContentType `json:"Content-Type"`
	Namespace   string                                   `json:"namespace"`
	Repository  string                                   `json:"repository"`
	Offset      *string                                  `json:"offset,omitempty"`
	Limit       *string                                  `json:"limit,omitempty"`
}

func (o ListRetentionHistoriesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListRetentionHistoriesRequest struct{}"
	}

	return strings.Join([]string{"ListRetentionHistoriesRequest", string(data)}, " ")
}

type ListRetentionHistoriesRequestContentType struct {
	value string
}

type ListRetentionHistoriesRequestContentTypeEnum struct {
	APPLICATION_JSONCHARSETUTF_8 ListRetentionHistoriesRequestContentType
	APPLICATION_JSON             ListRetentionHistoriesRequestContentType
}

func GetListRetentionHistoriesRequestContentTypeEnum() ListRetentionHistoriesRequestContentTypeEnum {
	return ListRetentionHistoriesRequestContentTypeEnum{
		APPLICATION_JSONCHARSETUTF_8: ListRetentionHistoriesRequestContentType{
			value: "application/json;charset=utf-8",
		},
		APPLICATION_JSON: ListRetentionHistoriesRequestContentType{
			value: "application/json",
		},
	}
}

func (c ListRetentionHistoriesRequestContentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListRetentionHistoriesRequestContentType) UnmarshalJSON(b []byte) error {
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

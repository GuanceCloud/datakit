/*
 * kms
 *
 * KMS v1.0 API, open API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// Response Object
type ListGrantsResponse struct {
	// grant列表，详情请参见grants字段数据结构说明。
	Grants *[]Grants `json:"grants,omitempty"`
	// 获取下一页所需要传递的marker值。 当“truncated”为“false”时，“next_marker”为空。
	NextMarker *string `json:"next_marker,omitempty"`
	// 是否还有下一页：  - “true”表示还有数据。  - “false”表示已经是最后一页。
	Truncated *ListGrantsResponseTruncated `json:"truncated,omitempty"`
	// grant总条数。
	Total          *int32 `json:"total,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListGrantsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListGrantsResponse struct{}"
	}

	return strings.Join([]string{"ListGrantsResponse", string(data)}, " ")
}

type ListGrantsResponseTruncated struct {
	value string
}

type ListGrantsResponseTruncatedEnum struct {
	TRUE  ListGrantsResponseTruncated
	FALSE ListGrantsResponseTruncated
}

func GetListGrantsResponseTruncatedEnum() ListGrantsResponseTruncatedEnum {
	return ListGrantsResponseTruncatedEnum{
		TRUE: ListGrantsResponseTruncated{
			value: "true",
		},
		FALSE: ListGrantsResponseTruncated{
			value: "false",
		},
	}
}

func (c ListGrantsResponseTruncated) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListGrantsResponseTruncated) UnmarshalJSON(b []byte) error {
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

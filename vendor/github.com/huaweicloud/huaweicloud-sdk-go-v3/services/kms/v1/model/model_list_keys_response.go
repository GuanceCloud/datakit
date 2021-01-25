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
type ListKeysResponse struct {
	// key_id列表。
	Keys *[]string `json:"keys,omitempty"`
	// 密钥详情列表。详情参见KeyDetails
	KeyDetails *[]KeyDetails `json:"key_details,omitempty"`
	// 获取下一页所需要传递的“marker”值。当“truncated”为“false”时，“next_marker”为空。
	NextMarker *string `json:"next_marker,omitempty"`
	// 是否还有下一页： - “true”表示还有数据。 - “false”表示已经是最后一页。
	Truncated *ListKeysResponseTruncated `json:"truncated,omitempty"`
	// 密钥总条数。
	Total          *int32 `json:"total,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListKeysResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListKeysResponse struct{}"
	}

	return strings.Join([]string{"ListKeysResponse", string(data)}, " ")
}

type ListKeysResponseTruncated struct {
	value string
}

type ListKeysResponseTruncatedEnum struct {
	TRUE  ListKeysResponseTruncated
	FALSE ListKeysResponseTruncated
}

func GetListKeysResponseTruncatedEnum() ListKeysResponseTruncatedEnum {
	return ListKeysResponseTruncatedEnum{
		TRUE: ListKeysResponseTruncated{
			value: "true",
		},
		FALSE: ListKeysResponseTruncated{
			value: "false",
		},
	}
}

func (c ListKeysResponseTruncated) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListKeysResponseTruncated) UnmarshalJSON(b []byte) error {
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

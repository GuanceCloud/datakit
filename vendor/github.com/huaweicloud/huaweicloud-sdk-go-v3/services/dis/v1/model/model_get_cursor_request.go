/*
 * DIS
 *
 * DIS v1 API
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
type GetCursorRequest struct {
	StreamName             string                      `json:"stream-name"`
	PartitionId            string                      `json:"partition-id"`
	CursorType             *GetCursorRequestCursorType `json:"cursor-type,omitempty"`
	StartingSequenceNumber *string                     `json:"starting-sequence-number,omitempty"`
	Timestamp              *int64                      `json:"timestamp,omitempty"`
}

func (o GetCursorRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "GetCursorRequest struct{}"
	}

	return strings.Join([]string{"GetCursorRequest", string(data)}, " ")
}

type GetCursorRequestCursorType struct {
	value string
}

type GetCursorRequestCursorTypeEnum struct {
	AT_SEQUENCE_NUMBER    GetCursorRequestCursorType
	AFTER_SEQUENCE_NUMBER GetCursorRequestCursorType
	TRIM_HORIZON          GetCursorRequestCursorType
	LATEST                GetCursorRequestCursorType
	AT_TIMESTAMP          GetCursorRequestCursorType
}

func GetGetCursorRequestCursorTypeEnum() GetCursorRequestCursorTypeEnum {
	return GetCursorRequestCursorTypeEnum{
		AT_SEQUENCE_NUMBER: GetCursorRequestCursorType{
			value: "AT_SEQUENCE_NUMBER",
		},
		AFTER_SEQUENCE_NUMBER: GetCursorRequestCursorType{
			value: "AFTER_SEQUENCE_NUMBER",
		},
		TRIM_HORIZON: GetCursorRequestCursorType{
			value: "TRIM_HORIZON",
		},
		LATEST: GetCursorRequestCursorType{
			value: "LATEST",
		},
		AT_TIMESTAMP: GetCursorRequestCursorType{
			value: "AT_TIMESTAMP",
		},
	}
}

func (c GetCursorRequestCursorType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *GetCursorRequestCursorType) UnmarshalJSON(b []byte) error {
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

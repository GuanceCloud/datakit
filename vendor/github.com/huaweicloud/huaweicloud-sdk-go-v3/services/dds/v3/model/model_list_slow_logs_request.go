/*
 * DDS
 *
 * API v3
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
type ListSlowLogsRequest struct {
	InstanceId string                   `json:"instance_id"`
	StartDate  string                   `json:"start_date"`
	EndDate    string                   `json:"end_date"`
	NodeId     *string                  `json:"node_id,omitempty"`
	Type       *ListSlowLogsRequestType `json:"type,omitempty"`
	Offset     *int32                   `json:"offset,omitempty"`
	Limit      *int32                   `json:"limit,omitempty"`
}

func (o ListSlowLogsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListSlowLogsRequest struct{}"
	}

	return strings.Join([]string{"ListSlowLogsRequest", string(data)}, " ")
}

type ListSlowLogsRequestType struct {
	value string
}

type ListSlowLogsRequestTypeEnum struct {
	INSERT      ListSlowLogsRequestType
	QUERY       ListSlowLogsRequestType
	UPDATE      ListSlowLogsRequestType
	REMOVE      ListSlowLogsRequestType
	GETMORE     ListSlowLogsRequestType
	COMMAND     ListSlowLogsRequestType
	KILLCURSORS ListSlowLogsRequestType
}

func GetListSlowLogsRequestTypeEnum() ListSlowLogsRequestTypeEnum {
	return ListSlowLogsRequestTypeEnum{
		INSERT: ListSlowLogsRequestType{
			value: "INSERT",
		},
		QUERY: ListSlowLogsRequestType{
			value: "QUERY",
		},
		UPDATE: ListSlowLogsRequestType{
			value: "UPDATE",
		},
		REMOVE: ListSlowLogsRequestType{
			value: "REMOVE",
		},
		GETMORE: ListSlowLogsRequestType{
			value: "GETMORE",
		},
		COMMAND: ListSlowLogsRequestType{
			value: "COMMAND",
		},
		KILLCURSORS: ListSlowLogsRequestType{
			value: "KILLCURSORS",
		},
	}
}

func (c ListSlowLogsRequestType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListSlowLogsRequestType) UnmarshalJSON(b []byte) error {
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

/*
 * RDS
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
	XLanguage  *string                  `json:"X-Language,omitempty"`
	InstanceId string                   `json:"instance_id"`
	StartDate  string                   `json:"start_date"`
	EndDate    string                   `json:"end_date"`
	Offset     *int32                   `json:"offset,omitempty"`
	Limit      *int32                   `json:"limit,omitempty"`
	Type       *ListSlowLogsRequestType `json:"type,omitempty"`
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
	INSERT ListSlowLogsRequestType
	UPDATE ListSlowLogsRequestType
	SELECT ListSlowLogsRequestType
	DELETE ListSlowLogsRequestType
	CREATE ListSlowLogsRequestType
}

func GetListSlowLogsRequestTypeEnum() ListSlowLogsRequestTypeEnum {
	return ListSlowLogsRequestTypeEnum{
		INSERT: ListSlowLogsRequestType{
			value: "INSERT",
		},
		UPDATE: ListSlowLogsRequestType{
			value: "UPDATE",
		},
		SELECT: ListSlowLogsRequestType{
			value: "SELECT",
		},
		DELETE: ListSlowLogsRequestType{
			value: "DELETE",
		},
		CREATE: ListSlowLogsRequestType{
			value: "CREATE",
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

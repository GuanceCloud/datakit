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
type ListErrorLogsRequest struct {
	XLanguage  *string                    `json:"X-Language,omitempty"`
	InstanceId string                     `json:"instance_id"`
	StartDate  string                     `json:"start_date"`
	EndDate    string                     `json:"end_date"`
	Offset     *int32                     `json:"offset,omitempty"`
	Limit      *int32                     `json:"limit,omitempty"`
	Level      *ListErrorLogsRequestLevel `json:"level,omitempty"`
}

func (o ListErrorLogsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListErrorLogsRequest struct{}"
	}

	return strings.Join([]string{"ListErrorLogsRequest", string(data)}, " ")
}

type ListErrorLogsRequestLevel struct {
	value string
}

type ListErrorLogsRequestLevelEnum struct {
	ALL     ListErrorLogsRequestLevel
	INFO    ListErrorLogsRequestLevel
	LOG     ListErrorLogsRequestLevel
	WARNING ListErrorLogsRequestLevel
	ERROR   ListErrorLogsRequestLevel
	FATAL   ListErrorLogsRequestLevel
	PANIC   ListErrorLogsRequestLevel
	NOTE    ListErrorLogsRequestLevel
}

func GetListErrorLogsRequestLevelEnum() ListErrorLogsRequestLevelEnum {
	return ListErrorLogsRequestLevelEnum{
		ALL: ListErrorLogsRequestLevel{
			value: "ALL",
		},
		INFO: ListErrorLogsRequestLevel{
			value: "INFO",
		},
		LOG: ListErrorLogsRequestLevel{
			value: "LOG",
		},
		WARNING: ListErrorLogsRequestLevel{
			value: "WARNING",
		},
		ERROR: ListErrorLogsRequestLevel{
			value: "ERROR",
		},
		FATAL: ListErrorLogsRequestLevel{
			value: "FATAL",
		},
		PANIC: ListErrorLogsRequestLevel{
			value: "PANIC",
		},
		NOTE: ListErrorLogsRequestLevel{
			value: "NOTE",
		},
	}
}

func (c ListErrorLogsRequestLevel) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListErrorLogsRequestLevel) UnmarshalJSON(b []byte) error {
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

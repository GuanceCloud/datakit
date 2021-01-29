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
type ListSlowlogStatisticsRequest struct {
	XLanguage  *ListSlowlogStatisticsRequestXLanguage `json:"X-Language,omitempty"`
	InstanceId string                                 `json:"instance_id"`
	CurPage    int32                                  `json:"cur_page"`
	PerPage    int32                                  `json:"per_page"`
	StartDate  string                                 `json:"start_date"`
	EndDate    string                                 `json:"end_date"`
	Type       ListSlowlogStatisticsRequestType       `json:"type"`
}

func (o ListSlowlogStatisticsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListSlowlogStatisticsRequest struct{}"
	}

	return strings.Join([]string{"ListSlowlogStatisticsRequest", string(data)}, " ")
}

type ListSlowlogStatisticsRequestXLanguage struct {
	value string
}

type ListSlowlogStatisticsRequestXLanguageEnum struct {
	ZH_CN ListSlowlogStatisticsRequestXLanguage
	EN_US ListSlowlogStatisticsRequestXLanguage
}

func GetListSlowlogStatisticsRequestXLanguageEnum() ListSlowlogStatisticsRequestXLanguageEnum {
	return ListSlowlogStatisticsRequestXLanguageEnum{
		ZH_CN: ListSlowlogStatisticsRequestXLanguage{
			value: "zh-cn",
		},
		EN_US: ListSlowlogStatisticsRequestXLanguage{
			value: "en-us",
		},
	}
}

func (c ListSlowlogStatisticsRequestXLanguage) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListSlowlogStatisticsRequestXLanguage) UnmarshalJSON(b []byte) error {
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

type ListSlowlogStatisticsRequestType struct {
	value string
}

type ListSlowlogStatisticsRequestTypeEnum struct {
	INSERT ListSlowlogStatisticsRequestType
	UPDATE ListSlowlogStatisticsRequestType
	SELECT ListSlowlogStatisticsRequestType
	DELETE ListSlowlogStatisticsRequestType
	CREATE ListSlowlogStatisticsRequestType
}

func GetListSlowlogStatisticsRequestTypeEnum() ListSlowlogStatisticsRequestTypeEnum {
	return ListSlowlogStatisticsRequestTypeEnum{
		INSERT: ListSlowlogStatisticsRequestType{
			value: "INSERT",
		},
		UPDATE: ListSlowlogStatisticsRequestType{
			value: "UPDATE",
		},
		SELECT: ListSlowlogStatisticsRequestType{
			value: "SELECT",
		},
		DELETE: ListSlowlogStatisticsRequestType{
			value: "DELETE",
		},
		CREATE: ListSlowlogStatisticsRequestType{
			value: "CREATE",
		},
	}
}

func (c ListSlowlogStatisticsRequestType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListSlowlogStatisticsRequestType) UnmarshalJSON(b []byte) error {
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

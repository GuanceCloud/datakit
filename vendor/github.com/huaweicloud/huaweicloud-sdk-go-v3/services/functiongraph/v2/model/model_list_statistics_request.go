/*
 * FunctionGraph
 *
 * API v2
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
type ListStatisticsRequest struct {
	Filter    ListStatisticsRequestFilter    `json:"filter"`
	Period    *string                        `json:"period,omitempty"`
	MonthCode ListStatisticsRequestMonthCode `json:"month_code"`
}

func (o ListStatisticsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListStatisticsRequest struct{}"
	}

	return strings.Join([]string{"ListStatisticsRequest", string(data)}, " ")
}

type ListStatisticsRequestFilter struct {
	value string
}

type ListStatisticsRequestFilterEnum struct {
	MONTHLY_STATISTICS ListStatisticsRequestFilter
	METRIC             ListStatisticsRequestFilter
}

func GetListStatisticsRequestFilterEnum() ListStatisticsRequestFilterEnum {
	return ListStatisticsRequestFilterEnum{
		MONTHLY_STATISTICS: ListStatisticsRequestFilter{
			value: "monthly_statistics",
		},
		METRIC: ListStatisticsRequestFilter{
			value: "metric",
		},
	}
}

func (c ListStatisticsRequestFilter) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListStatisticsRequestFilter) UnmarshalJSON(b []byte) error {
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

type ListStatisticsRequestMonthCode struct {
	value string
}

type ListStatisticsRequestMonthCodeEnum struct {
	E_0 ListStatisticsRequestMonthCode
	E_1 ListStatisticsRequestMonthCode
	E_2 ListStatisticsRequestMonthCode
	E_3 ListStatisticsRequestMonthCode
}

func GetListStatisticsRequestMonthCodeEnum() ListStatisticsRequestMonthCodeEnum {
	return ListStatisticsRequestMonthCodeEnum{
		E_0: ListStatisticsRequestMonthCode{
			value: "0",
		},
		E_1: ListStatisticsRequestMonthCode{
			value: "1",
		},
		E_2: ListStatisticsRequestMonthCode{
			value: "2",
		},
		E_3: ListStatisticsRequestMonthCode{
			value: "3",
		},
	}
}

func (c ListStatisticsRequestMonthCode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListStatisticsRequestMonthCode) UnmarshalJSON(b []byte) error {
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

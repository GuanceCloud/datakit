/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// 单个大key分析历史记录
type RecordsResponse struct {
	// 大key分析记录ID
	Id *string `json:"id,omitempty"`
	// 分析任务状态
	Status *RecordsResponseStatus `json:"status,omitempty"`
	// 分析方式
	ScanType *RecordsResponseScanType `json:"scan_type,omitempty"`
	// 分析任务创建时间,格式为：\"2020-06-15T02:21:18.669Z\"
	CreatedAt *string `json:"created_at,omitempty"`
	// 分析任务开始时间,格式为：\"2020-06-15T02:21:18.669Z\"
	StartedAt *string `json:"started_at,omitempty"`
	// 分析任务结束时间,格式为：\"2020-06-15T02:21:18.669Z\"
	FinishedAt *string `json:"finished_at,omitempty"`
}

func (o RecordsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RecordsResponse struct{}"
	}

	return strings.Join([]string{"RecordsResponse", string(data)}, " ")
}

type RecordsResponseStatus struct {
	value string
}

type RecordsResponseStatusEnum struct {
	WAITING RecordsResponseStatus
	RUNNING RecordsResponseStatus
	SUCCESS RecordsResponseStatus
	FAILED  RecordsResponseStatus
}

func GetRecordsResponseStatusEnum() RecordsResponseStatusEnum {
	return RecordsResponseStatusEnum{
		WAITING: RecordsResponseStatus{
			value: "waiting",
		},
		RUNNING: RecordsResponseStatus{
			value: "running",
		},
		SUCCESS: RecordsResponseStatus{
			value: "success",
		},
		FAILED: RecordsResponseStatus{
			value: "failed",
		},
	}
}

func (c RecordsResponseStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *RecordsResponseStatus) UnmarshalJSON(b []byte) error {
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

type RecordsResponseScanType struct {
	value string
}

type RecordsResponseScanTypeEnum struct {
	MANUAL RecordsResponseScanType
	AUTO   RecordsResponseScanType
}

func GetRecordsResponseScanTypeEnum() RecordsResponseScanTypeEnum {
	return RecordsResponseScanTypeEnum{
		MANUAL: RecordsResponseScanType{
			value: "manual",
		},
		AUTO: RecordsResponseScanType{
			value: "auto",
		},
	}
}

func (c RecordsResponseScanType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *RecordsResponseScanType) UnmarshalJSON(b []byte) error {
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

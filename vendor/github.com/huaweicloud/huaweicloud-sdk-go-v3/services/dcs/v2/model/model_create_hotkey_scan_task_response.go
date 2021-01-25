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

// Response Object
type CreateHotkeyScanTaskResponse struct {
	// 热key分析记录ID
	Id *string `json:"id,omitempty"`
	// 实例ID
	InstanceId *string `json:"instance_id,omitempty"`
	// 分析任务状态
	Status *CreateHotkeyScanTaskResponseStatus `json:"status,omitempty"`
	// 分析方式
	ScanType *CreateHotkeyScanTaskResponseScanType `json:"scan_type,omitempty"`
	// 分析任务创建时间,格式为：\"2020-06-15T02:21:18.669Z\"
	CreatedAt *string `json:"created_at,omitempty"`
	// 分析任务开始时间,格式为：\"2020-06-15T02:21:18.669Z\"
	StartedAt *string `json:"started_at,omitempty"`
	// 分析任务结束时间,格式为：\"2020-06-15T02:21:18.669Z\"
	FinishedAt *string `json:"finished_at,omitempty"`
	// 热key的数量
	Num *int32 `json:"num,omitempty"`
	// 热key记录
	Keys           *[]HotkeysBody `json:"keys,omitempty"`
	HttpStatusCode int            `json:"-"`
}

func (o CreateHotkeyScanTaskResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateHotkeyScanTaskResponse struct{}"
	}

	return strings.Join([]string{"CreateHotkeyScanTaskResponse", string(data)}, " ")
}

type CreateHotkeyScanTaskResponseStatus struct {
	value string
}

type CreateHotkeyScanTaskResponseStatusEnum struct {
	WAITING CreateHotkeyScanTaskResponseStatus
	RUNNING CreateHotkeyScanTaskResponseStatus
	SUCCESS CreateHotkeyScanTaskResponseStatus
	FAILED  CreateHotkeyScanTaskResponseStatus
}

func GetCreateHotkeyScanTaskResponseStatusEnum() CreateHotkeyScanTaskResponseStatusEnum {
	return CreateHotkeyScanTaskResponseStatusEnum{
		WAITING: CreateHotkeyScanTaskResponseStatus{
			value: "waiting",
		},
		RUNNING: CreateHotkeyScanTaskResponseStatus{
			value: "running",
		},
		SUCCESS: CreateHotkeyScanTaskResponseStatus{
			value: "success",
		},
		FAILED: CreateHotkeyScanTaskResponseStatus{
			value: "failed",
		},
	}
}

func (c CreateHotkeyScanTaskResponseStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateHotkeyScanTaskResponseStatus) UnmarshalJSON(b []byte) error {
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

type CreateHotkeyScanTaskResponseScanType struct {
	value string
}

type CreateHotkeyScanTaskResponseScanTypeEnum struct {
	MANUAL CreateHotkeyScanTaskResponseScanType
	AUTO   CreateHotkeyScanTaskResponseScanType
}

func GetCreateHotkeyScanTaskResponseScanTypeEnum() CreateHotkeyScanTaskResponseScanTypeEnum {
	return CreateHotkeyScanTaskResponseScanTypeEnum{
		MANUAL: CreateHotkeyScanTaskResponseScanType{
			value: "manual",
		},
		AUTO: CreateHotkeyScanTaskResponseScanType{
			value: "auto",
		},
	}
}

func (c CreateHotkeyScanTaskResponseScanType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateHotkeyScanTaskResponseScanType) UnmarshalJSON(b []byte) error {
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

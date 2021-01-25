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
type CreateBigkeyScanTaskResponse struct {
	// 大key分析记录ID
	Id *string `json:"id,omitempty"`
	// 实例ID
	InstanceId *string `json:"instance_id,omitempty"`
	// 分析任务状态
	Status *CreateBigkeyScanTaskResponseStatus `json:"status,omitempty"`
	// 分析方式
	ScanType *CreateBigkeyScanTaskResponseScanType `json:"scan_type,omitempty"`
	// 分析任务创建时间,格式为：\"2020-06-15T02:21:18.669Z\"
	CreatedAt *string `json:"created_at,omitempty"`
	// 分析任务开始时间,格式为：\"2020-06-15T02:21:18.669Z\"
	StartedAt *string `json:"started_at,omitempty"`
	// 分析任务结束时间,格式为：\"2020-06-15T02:21:18.669Z\"
	FinishedAt *string `json:"finished_at,omitempty"`
	// 大key的数量
	Num *int32 `json:"num,omitempty"`
	// 大key记录
	Keys           *[]BigkeysBody `json:"keys,omitempty"`
	HttpStatusCode int            `json:"-"`
}

func (o CreateBigkeyScanTaskResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateBigkeyScanTaskResponse struct{}"
	}

	return strings.Join([]string{"CreateBigkeyScanTaskResponse", string(data)}, " ")
}

type CreateBigkeyScanTaskResponseStatus struct {
	value string
}

type CreateBigkeyScanTaskResponseStatusEnum struct {
	WAITING CreateBigkeyScanTaskResponseStatus
	RUNNING CreateBigkeyScanTaskResponseStatus
	SUCCESS CreateBigkeyScanTaskResponseStatus
	FAILED  CreateBigkeyScanTaskResponseStatus
}

func GetCreateBigkeyScanTaskResponseStatusEnum() CreateBigkeyScanTaskResponseStatusEnum {
	return CreateBigkeyScanTaskResponseStatusEnum{
		WAITING: CreateBigkeyScanTaskResponseStatus{
			value: "waiting",
		},
		RUNNING: CreateBigkeyScanTaskResponseStatus{
			value: "running",
		},
		SUCCESS: CreateBigkeyScanTaskResponseStatus{
			value: "success",
		},
		FAILED: CreateBigkeyScanTaskResponseStatus{
			value: "failed",
		},
	}
}

func (c CreateBigkeyScanTaskResponseStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateBigkeyScanTaskResponseStatus) UnmarshalJSON(b []byte) error {
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

type CreateBigkeyScanTaskResponseScanType struct {
	value string
}

type CreateBigkeyScanTaskResponseScanTypeEnum struct {
	MANUAL CreateBigkeyScanTaskResponseScanType
	AUTO   CreateBigkeyScanTaskResponseScanType
}

func GetCreateBigkeyScanTaskResponseScanTypeEnum() CreateBigkeyScanTaskResponseScanTypeEnum {
	return CreateBigkeyScanTaskResponseScanTypeEnum{
		MANUAL: CreateBigkeyScanTaskResponseScanType{
			value: "manual",
		},
		AUTO: CreateBigkeyScanTaskResponseScanType{
			value: "auto",
		},
	}
}

func (c CreateBigkeyScanTaskResponseScanType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateBigkeyScanTaskResponseScanType) UnmarshalJSON(b []byte) error {
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

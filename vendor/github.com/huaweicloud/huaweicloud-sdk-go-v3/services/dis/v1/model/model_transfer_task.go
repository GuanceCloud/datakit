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

type TransferTask struct {
	// 转储任务名称。
	TaskName *string `json:"task_name,omitempty"`
	// 转储任务状态。  - ERROR：错误。 - STARTING：启动中。 - PAUSED：已停止。 - RUNNING：运行中。 - DELETE：已删除。 - ABNORMAL：异常。
	State *TransferTaskState `json:"state,omitempty"`
	// 转储任务类型。  - OBS：转储到OBS。 - MRS：转储到MRS。 - DLI：转储到DLI。 - CLOUDTABLE：转储到CloudTable。 - DWS：转储到DWS。
	DestinationType *TransferTaskDestinationType `json:"destination_type,omitempty"`
	// 转储任务创建时间。
	CreateTime *int64 `json:"create_time,omitempty"`
	// 转储任务最近一次转储时间。
	LastTransferTimestamp *int64 `json:"last_transfer_timestamp,omitempty"`
}

func (o TransferTask) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "TransferTask struct{}"
	}

	return strings.Join([]string{"TransferTask", string(data)}, " ")
}

type TransferTaskState struct {
	value string
}

type TransferTaskStateEnum struct {
	ERROR    TransferTaskState
	STARTING TransferTaskState
	PAUSED   TransferTaskState
	RUNNING  TransferTaskState
	DELETE   TransferTaskState
	ABNORMAL TransferTaskState
}

func GetTransferTaskStateEnum() TransferTaskStateEnum {
	return TransferTaskStateEnum{
		ERROR: TransferTaskState{
			value: "ERROR",
		},
		STARTING: TransferTaskState{
			value: "STARTING",
		},
		PAUSED: TransferTaskState{
			value: "PAUSED",
		},
		RUNNING: TransferTaskState{
			value: "RUNNING",
		},
		DELETE: TransferTaskState{
			value: "DELETE",
		},
		ABNORMAL: TransferTaskState{
			value: "ABNORMAL",
		},
	}
}

func (c TransferTaskState) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *TransferTaskState) UnmarshalJSON(b []byte) error {
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

type TransferTaskDestinationType struct {
	value string
}

type TransferTaskDestinationTypeEnum struct {
	OBS        TransferTaskDestinationType
	MRS        TransferTaskDestinationType
	DLI        TransferTaskDestinationType
	CLOUDTABLE TransferTaskDestinationType
	DWS        TransferTaskDestinationType
}

func GetTransferTaskDestinationTypeEnum() TransferTaskDestinationTypeEnum {
	return TransferTaskDestinationTypeEnum{
		OBS: TransferTaskDestinationType{
			value: "OBS",
		},
		MRS: TransferTaskDestinationType{
			value: "MRS",
		},
		DLI: TransferTaskDestinationType{
			value: "DLI",
		},
		CLOUDTABLE: TransferTaskDestinationType{
			value: "CLOUDTABLE",
		},
		DWS: TransferTaskDestinationType{
			value: "DWS",
		},
	}
}

func (c TransferTaskDestinationType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *TransferTaskDestinationType) UnmarshalJSON(b []byte) error {
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

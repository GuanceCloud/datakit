/*
 * kps
 *
 * kps v3 版本API
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
type ListKeypairTaskResponse struct {
	// 租户虚拟机ID
	ServerId *string `json:"server_id,omitempty"`
	// 任务下发成功返回的ID
	TaskId *string `json:"task_id,omitempty"`
	// 密钥对正在处理的状态。 - READY_RESET 准备重置 - RUNNING_RESET 正在重置 - FAILED_RESET 重置失败 - SUCCESS_RESET 重置成功 - READY_REPLACE 准备替换 - RUNNING_REPLACE 正在替换 - FAILED_RESET 替换失败 - SUCCESS_RESET 替换成功 - READY_UNBIND 准备解绑 - RUNNING_UNBIND 正在解绑 - FAILED_UNBIND 解绑失败 - SUCCESS_UNBIND 解绑成功
	TaskStatus     *ListKeypairTaskResponseTaskStatus `json:"task_status,omitempty"`
	HttpStatusCode int                                `json:"-"`
}

func (o ListKeypairTaskResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListKeypairTaskResponse struct{}"
	}

	return strings.Join([]string{"ListKeypairTaskResponse", string(data)}, " ")
}

type ListKeypairTaskResponseTaskStatus struct {
	value string
}

type ListKeypairTaskResponseTaskStatusEnum struct {
	READY_RESET     ListKeypairTaskResponseTaskStatus
	RUNNING_RESET   ListKeypairTaskResponseTaskStatus
	FAILED_RESET    ListKeypairTaskResponseTaskStatus
	SUCCESS_RESET   ListKeypairTaskResponseTaskStatus
	READY_REPLACE   ListKeypairTaskResponseTaskStatus
	RUNNING_REPLACE ListKeypairTaskResponseTaskStatus
	READY_UNBIND    ListKeypairTaskResponseTaskStatus
	RUNNING_UNBIND  ListKeypairTaskResponseTaskStatus
	FAILED_UNBIND   ListKeypairTaskResponseTaskStatus
	SUCCESS_UNBIND  ListKeypairTaskResponseTaskStatus
}

func GetListKeypairTaskResponseTaskStatusEnum() ListKeypairTaskResponseTaskStatusEnum {
	return ListKeypairTaskResponseTaskStatusEnum{
		READY_RESET: ListKeypairTaskResponseTaskStatus{
			value: "READY_RESET",
		},
		RUNNING_RESET: ListKeypairTaskResponseTaskStatus{
			value: "RUNNING_RESET",
		},
		FAILED_RESET: ListKeypairTaskResponseTaskStatus{
			value: "FAILED_RESET",
		},
		SUCCESS_RESET: ListKeypairTaskResponseTaskStatus{
			value: "SUCCESS_RESET",
		},
		READY_REPLACE: ListKeypairTaskResponseTaskStatus{
			value: "READY_REPLACE",
		},
		RUNNING_REPLACE: ListKeypairTaskResponseTaskStatus{
			value: "RUNNING_REPLACE",
		},
		READY_UNBIND: ListKeypairTaskResponseTaskStatus{
			value: "READY_UNBIND",
		},
		RUNNING_UNBIND: ListKeypairTaskResponseTaskStatus{
			value: "RUNNING_UNBIND",
		},
		FAILED_UNBIND: ListKeypairTaskResponseTaskStatus{
			value: "FAILED_UNBIND",
		},
		SUCCESS_UNBIND: ListKeypairTaskResponseTaskStatus{
			value: "SUCCESS_UNBIND",
		},
	}
}

func (c ListKeypairTaskResponseTaskStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListKeypairTaskResponseTaskStatus) UnmarshalJSON(b []byte) error {
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

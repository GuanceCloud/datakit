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

// 失败的任务详细信息。
type FailedTasks struct {
	// 虚拟机ID
	TaskId *string `json:"task_id,omitempty"`
	// 任务的操作类型。 - FAILED_RESET 重置 - FAILED_REPLACE 替换 - FAILED_UNBIND 解绑
	OperateType *FailedTasksOperateType `json:"operate_type,omitempty"`
	// 任务时间
	TaskTime *int64 `json:"task_time,omitempty"`
	// 任务失败错误码
	TaskErrorCode *string `json:"task_error_code,omitempty"`
	// 任务失败错误码
	TaskErrorMsg *string `json:"task_error_msg,omitempty"`
	// 虚拟机名称
	ServerName *string `json:"server_name,omitempty"`
	// 虚拟机ID
	ServerId *string `json:"server_id,omitempty"`
	// 密钥对名称
	KeypairName *string `json:"keypair_name,omitempty"`
}

func (o FailedTasks) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "FailedTasks struct{}"
	}

	return strings.Join([]string{"FailedTasks", string(data)}, " ")
}

type FailedTasksOperateType struct {
	value string
}

type FailedTasksOperateTypeEnum struct {
	FAILED_RESET   FailedTasksOperateType
	FAILED_REPLACE FailedTasksOperateType
	FAILED_UNBIND  FailedTasksOperateType
}

func GetFailedTasksOperateTypeEnum() FailedTasksOperateTypeEnum {
	return FailedTasksOperateTypeEnum{
		FAILED_RESET: FailedTasksOperateType{
			value: "FAILED_RESET",
		},
		FAILED_REPLACE: FailedTasksOperateType{
			value: "FAILED_REPLACE",
		},
		FAILED_UNBIND: FailedTasksOperateType{
			value: "FAILED_UNBIND",
		},
	}
}

func (c FailedTasksOperateType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *FailedTasksOperateType) UnmarshalJSON(b []byte) error {
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

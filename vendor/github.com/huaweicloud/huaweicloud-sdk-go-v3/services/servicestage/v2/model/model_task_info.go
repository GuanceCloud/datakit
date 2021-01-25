/*
 * ServiceStage
 *
 * ServiceStage的API,包括应用管理和仓库授权管理
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// 。
type TaskInfo struct {
	// 创建时间。
	CreatedAt *string `json:"CREATED_AT,omitempty"`
	// 健康检查时间。
	LastHealthCheck *string `json:"LAST_HEALTH_CHECK,omitempty"`
	// 消息。
	Messages *string `json:"MESSAGES,omitempty"`
	// 创建用户ID。
	OwnerId *string `json:"OWNER_ID,omitempty"`
	// 任务ID。
	TaskId *string `json:"TASK_ID,omitempty"`
	// 任务序号。
	TaskIndex *int32 `json:"TASK_INDEX,omitempty"`
	// 任务名称。
	TaskName *string `json:"TASK_NAME,omitempty"`
	// 任务状态。
	TaskStatus *TaskInfoTaskStatus `json:"TASK_STATUS,omitempty"`
	// 任务类型。
	TaskType *string `json:"TASK_TYPE,omitempty"`
}

func (o TaskInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "TaskInfo struct{}"
	}

	return strings.Join([]string{"TaskInfo", string(data)}, " ")
}

type TaskInfoTaskStatus struct {
	value string
}

type TaskInfoTaskStatusEnum struct {
	RUNNING   TaskInfoTaskStatus
	SKIPPED   TaskInfoTaskStatus
	FAILED    TaskInfoTaskStatus
	SUCCEEDED TaskInfoTaskStatus
}

func GetTaskInfoTaskStatusEnum() TaskInfoTaskStatusEnum {
	return TaskInfoTaskStatusEnum{
		RUNNING: TaskInfoTaskStatus{
			value: "RUNNING",
		},
		SKIPPED: TaskInfoTaskStatus{
			value: "SKIPPED",
		},
		FAILED: TaskInfoTaskStatus{
			value: "FAILED",
		},
		SUCCEEDED: TaskInfoTaskStatus{
			value: "SUCCEEDED",
		},
	}
}

func (c TaskInfoTaskStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *TaskInfoTaskStatus) UnmarshalJSON(b []byte) error {
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

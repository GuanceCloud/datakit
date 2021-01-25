/*
 * Kafka
 *
 * Kafka Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type ListSinkTasksRespTasks struct {
	// 任务ID。
	TaskId *string `json:"task_id,omitempty"`
	// 转储任务名称。
	TaskName *string `json:"task_name,omitempty"`
	// 转储任务类型。
	DestinationType *string `json:"destination_type,omitempty"`
	// 转储任务创建时间戳。
	CreateTime *string `json:"create_time,omitempty"`
	// 转储任务状态。
	Status *string `json:"status,omitempty"`
	// 返回任务转存的topics列表或者正则表达式。
	Topics *string `json:"topics,omitempty"`
}

func (o ListSinkTasksRespTasks) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListSinkTasksRespTasks struct{}"
	}

	return strings.Join([]string{"ListSinkTasksRespTasks", string(data)}, " ")
}

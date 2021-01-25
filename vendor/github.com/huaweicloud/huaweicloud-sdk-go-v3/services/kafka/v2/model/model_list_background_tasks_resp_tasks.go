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

type ListBackgroundTasksRespTasks struct {
	// 任务ID。
	Id *string `json:"id,omitempty"`
	// 任务名称。
	Name *string `json:"name,omitempty"`
	// 用户名。
	UserName *string `json:"user_name,omitempty"`
	// 用户ID。
	UserId *string `json:"user_id,omitempty"`
	// 任务参数。
	Params *string `json:"params,omitempty"`
	// 任务状态。
	Status *string `json:"status,omitempty"`
	// 启动时间。
	CreatedAt *string `json:"created_at,omitempty"`
	// 结束时间。
	UpdatedAt *string `json:"updated_at,omitempty"`
}

func (o ListBackgroundTasksRespTasks) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListBackgroundTasksRespTasks struct{}"
	}

	return strings.Join([]string{"ListBackgroundTasksRespTasks", string(data)}, " ")
}

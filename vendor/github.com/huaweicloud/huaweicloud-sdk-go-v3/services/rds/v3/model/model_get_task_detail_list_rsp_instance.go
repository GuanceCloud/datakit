/*
 * RDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 执行任务的实例信息。
type GetTaskDetailListRspInstance struct {
	// 实例ID。
	Id string `json:"id"`
	// 实例名称。
	Name string `json:"name"`
}

func (o GetTaskDetailListRspInstance) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "GetTaskDetailListRspInstance struct{}"
	}

	return strings.Join([]string{"GetTaskDetailListRspInstance", string(data)}, " ")
}

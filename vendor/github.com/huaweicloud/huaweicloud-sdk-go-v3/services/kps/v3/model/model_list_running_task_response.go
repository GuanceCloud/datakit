/*
 * kps
 *
 * kps v3 版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListRunningTaskResponse struct {
	// 正在处理的任务总数。
	Total *int32 `json:"total,omitempty"`
	// 正在处理的任务列表。
	Tasks          *[]RunningTasks `json:"tasks,omitempty"`
	HttpStatusCode int             `json:"-"`
}

func (o ListRunningTaskResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListRunningTaskResponse struct{}"
	}

	return strings.Join([]string{"ListRunningTaskResponse", string(data)}, " ")
}

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

// Response Object
type ListJobInfoDetailResponse struct {
	Jobs *GetTaskDetailListRspJobs `json:"jobs,omitempty"`
	// 任务执行的具体的参数信息，为空则不返回该字段。
	TaskDetail *string                       `json:"task_detail,omitempty"`
	Instance   *GetTaskDetailListRspInstance `json:"instance,omitempty"`
	// 根据不同的任务，显示不同的内容。
	Entities *interface{} `json:"entities,omitempty"`
	// 任务执行失败时的错误信息。
	FailReason     *string `json:"fail_reason,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ListJobInfoDetailResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListJobInfoDetailResponse struct{}"
	}

	return strings.Join([]string{"ListJobInfoDetailResponse", string(data)}, " ")
}

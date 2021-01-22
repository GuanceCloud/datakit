/*
 * ProjectMan
 *
 * devcloud projectman api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 项目工作项概览信息
type IssueCompletionRateResponseV4 struct {
	IssueStatus *IssueStatusResponseV4 `json:"issue_status,omitempty"`
	// 工作项类型,2任务/task,3缺陷/bug,5epic,6feature,7story
	TrackerId *int32 `json:"tracker_id,omitempty"`
}

func (o IssueCompletionRateResponseV4) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "IssueCompletionRateResponseV4 struct{}"
	}

	return strings.Join([]string{"IssueCompletionRateResponseV4", string(data)}, " ")
}

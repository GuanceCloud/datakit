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

// 工作项不同状态下的数量
type IssueCompletionRateV4IssueStatus struct {
	// 已关闭的工作项
	ClosedNum *int32 `json:"closed_num,omitempty"`
	// 新建的工作项
	NewNum *int32 `json:"new_num,omitempty"`
	// 进行中的工作项数目
	ProcessNum *int32 `json:"process_num,omitempty"`
	// 已经拒绝的工作项
	RejectedNum *int32 `json:"rejected_num,omitempty"`
	// 已经解决的工作项
	SolvedNum *int32 `json:"solved_num,omitempty"`
	// 测试中的工作项
	TestNum *int32 `json:"test_num,omitempty"`
}

func (o IssueCompletionRateV4IssueStatus) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "IssueCompletionRateV4IssueStatus struct{}"
	}

	return strings.Join([]string{"IssueCompletionRateV4IssueStatus", string(data)}, " ")
}

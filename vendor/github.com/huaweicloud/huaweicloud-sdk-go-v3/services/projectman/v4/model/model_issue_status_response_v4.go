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

// 工作项的统计信息
type IssueStatusResponseV4 struct {
	// 已关闭数量
	ClosedNum *int32 `json:"closed_num,omitempty"`
	// 新建的数量
	NewNum *int32 `json:"new_num,omitempty"`
	// 开发中的数量
	ProcessNum *int32 `json:"process_num,omitempty"`
	// 已拒绝数量
	RejectedNum *int32 `json:"rejected_num,omitempty"`
	// 已解决数量
	SolvedNum *int32 `json:"solved_num,omitempty"`
	// 测试中的数量
	TestNum *int32 `json:"test_num,omitempty"`
}

func (o IssueStatusResponseV4) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "IssueStatusResponseV4 struct{}"
	}

	return strings.Join([]string{"IssueStatusResponseV4", string(data)}, " ")
}

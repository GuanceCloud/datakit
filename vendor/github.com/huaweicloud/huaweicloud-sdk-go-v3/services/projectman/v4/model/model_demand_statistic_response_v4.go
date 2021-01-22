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

// 需求概览信息
type DemandStatisticResponseV4 struct {
	// 已关闭数量
	ClosedNum *int32 `json:"closed_num,omitempty"`
	// 模块
	Module *string `json:"module,omitempty"`
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
	// 总数
	Total *int32 `json:"total,omitempty"`
}

func (o DemandStatisticResponseV4) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DemandStatisticResponseV4 struct{}"
	}

	return strings.Join([]string{"DemandStatisticResponseV4", string(data)}, " ")
}

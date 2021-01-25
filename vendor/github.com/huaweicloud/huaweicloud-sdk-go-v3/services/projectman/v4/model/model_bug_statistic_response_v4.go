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

// Bug信息
type BugStatisticResponseV4 struct {
	// 重要程度为关键的缺陷数
	CriticalNum *int32 `json:"critical_num,omitempty"`
	// DI
	DefectIndex *float64 `json:"defect_index,omitempty"`
	// 模块
	Module *string `json:"module,omitempty"`
	// 重要程度为一般的缺陷数
	NormalNum *int32 `json:"normal_num,omitempty"`
	// 重要程度为严重的缺陷数
	SeriousNum *int32 `json:"serious_num,omitempty"`
	// 重要程度为提示的缺陷数
	TipNum *int32 `json:"tip_num,omitempty"`
	// 总数
	Total *int32 `json:"total,omitempty"`
}

func (o BugStatisticResponseV4) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BugStatisticResponseV4 struct{}"
	}

	return strings.Join([]string{"BugStatisticResponseV4", string(data)}, " ")
}

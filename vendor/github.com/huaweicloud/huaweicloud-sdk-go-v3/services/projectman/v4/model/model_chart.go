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

type Chart struct {
	// 统计时间
	Date *string `json:"date,omitempty"`
	// 完成story工单
	FinishedNum *int32 `json:"finished_num,omitempty"`
	// 迭代id
	IterationId *int32 `json:"iteration_id,omitempty"`
	// 项目id
	ProjectNumId *int32 `json:"project_num_id,omitempty"`
	// 未完成story数
	RemainingNum *int32 `json:"remaining_num,omitempty"`
	// 总story数
	Total *int32 `json:"total,omitempty"`
}

func (o Chart) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Chart struct{}"
	}

	return strings.Join([]string{"Chart", string(data)}, " ")
}

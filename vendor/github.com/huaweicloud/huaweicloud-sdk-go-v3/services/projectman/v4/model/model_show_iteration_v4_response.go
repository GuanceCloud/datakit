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

// Response Object
type ShowIterationV4Response struct {
	// 迭代结束时间，年-月-日
	BeginTime *string `json:"begin_time,omitempty"`
	// 燃尽图
	Charts *[]Chart `json:"charts,omitempty"`
	// 已关闭的工单数
	ClosedTotal *int32 `json:"closed_total,omitempty"`
	// 迭代创建时间
	CreatedTime *string `json:"created_time,omitempty"`
	// 迭代开始时间，年-月-日
	EndTime *string `json:"end_time,omitempty"`
	// 是否有task
	HaveTask *bool `json:"have_task,omitempty"`
	// 迭代id
	IterationId *int32 `json:"iteration_id,omitempty"`
	// 迭代标题
	Name *string `json:"name,omitempty"`
	// 开启的工单数
	OpenedTotal *int32 `json:"opened_total,omitempty"`
	// 工作进展
	Progress *string `json:"progress,omitempty"`
	// 工单总数
	Total *int32 `json:"total,omitempty"`
	// 迭代更新时间
	UpdatedTime    *string `json:"updated_time,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ShowIterationV4Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowIterationV4Response struct{}"
	}

	return strings.Join([]string{"ShowIterationV4Response", string(data)}, " ")
}

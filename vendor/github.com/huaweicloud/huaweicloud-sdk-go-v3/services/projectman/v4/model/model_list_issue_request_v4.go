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

type ListIssueRequestV4 struct {
	// 处理人id
	AssignedIds *[]int32 `json:"assigned_ids,omitempty"`
	// 创建者id
	CreatorIds *[]int32 `json:"creator_ids,omitempty"`
	// 开发人id,对应用户信息的数字id
	DeveloperIds *[]int32 `json:"developer_ids,omitempty"`
	// id, 领域 14, '性能', 15, '功能', 16, '可靠性' 17, '网络安全' 18, '可维护性' 19, '其他DFX' 20, '可用性'
	DomainIds *[]int32 `json:"domain_ids,omitempty"`
	// 完成度
	DoneRatios *[]int32 `json:"done_ratios,omitempty"`
	// 迭代id
	IterationIds *[]int32 `json:"iteration_ids,omitempty"`
	// 每页显示数量
	Limit *int32 `json:"limit,omitempty"`
	// 模块id
	ModuleIds *[]int32 `json:"module_ids,omitempty"`
	// 分页索引，偏移量
	Offset *int32 `json:"offset,omitempty"`
	// 优先级
	PriorityIds *[]int32 `json:"priority_ids,omitempty"`
	// 查询类型 backlog feature epic
	QueryType *string `json:"query_type,omitempty"`
	// 查询类型
	SeverityIds *[]int32 `json:"severity_ids,omitempty"`
	// 状态   id 开始   1 进行中 2 已解决 3 测试中 4 已关闭 5 已解决 6
	StatusIds *[]int32 `json:"status_ids,omitempty"`
	// 故事点id
	StoryPointIds *[]int32 `json:"story_point_ids,omitempty"`
	// 工作项类型,2任务/task,3缺陷/bug,5epic,6feature,7story
	TrackerIds *[]int32 `json:"tracker_ids,omitempty"`
}

func (o ListIssueRequestV4) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListIssueRequestV4 struct{}"
	}

	return strings.Join([]string{"ListIssueRequestV4", string(data)}, " ")
}

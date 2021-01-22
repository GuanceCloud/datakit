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
type ShowProjectSummaryV4Response struct {
	// bug统计列表
	BugStatistics *[]BugStatisticResponseV4 `json:"bug_statistics,omitempty"`
	// 按模块统计列表
	DemandStatistics *[]DemandStatisticResponseV4 `json:"demand_statistics,omitempty"`
	// 按工作项类型统计列表
	IssueCompletionRates *[]IssueCompletionRateResponseV4 `json:"issue_completion_rates,omitempty"`
	// 项目id
	ProjectId      *string `json:"project_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ShowProjectSummaryV4Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowProjectSummaryV4Response struct{}"
	}

	return strings.Join([]string{"ShowProjectSummaryV4Response", string(data)}, " ")
}

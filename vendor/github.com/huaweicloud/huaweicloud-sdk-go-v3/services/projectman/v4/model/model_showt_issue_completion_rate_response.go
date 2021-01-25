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
type ShowtIssueCompletionRateResponse struct {
	// 不同类型的工作项完成率
	IssueCompletionRates *[]IssueCompletionRateV4IssueCompletionRates `json:"issue_completion_rates,omitempty"`
	// 总数
	Total          *int32 `json:"total,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ShowtIssueCompletionRateResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowtIssueCompletionRateResponse struct{}"
	}

	return strings.Join([]string{"ShowtIssueCompletionRateResponse", string(data)}, " ")
}

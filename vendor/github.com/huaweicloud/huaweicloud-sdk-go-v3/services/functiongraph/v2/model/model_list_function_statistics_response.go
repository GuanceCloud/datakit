/*
 * FunctionGraph
 *
 * API v2
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListFunctionStatisticsResponse struct {
	// 调用次数
	Count *[]SlaReportsValue `json:"count,omitempty"`
	// 平均时延，单位毫秒
	Duration *[]SlaReportsValue `json:"duration,omitempty"`
	// 错误次数
	FailCount *[]SlaReportsValue `json:"fail_count,omitempty"`
	// 最大时延，单位毫秒
	MaxDuration *[]SlaReportsValue `json:"max_duration,omitempty"`
	// 最小时延，单位毫秒
	MinDuration *[]SlaReportsValue `json:"min_duration,omitempty"`
	// 被拒绝次数
	RejectCount    *[]SlaReportsValue `json:"reject_count,omitempty"`
	HttpStatusCode int                `json:"-"`
}

func (o ListFunctionStatisticsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListFunctionStatisticsResponse struct{}"
	}

	return strings.Join([]string{"ListFunctionStatisticsResponse", string(data)}, " ")
}

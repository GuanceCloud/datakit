/*
 * RMS
 *
 * Resource Manager Api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ShowEvaluationStateByAssignmentIdResponse struct {
	// 规则ID
	PolicyAssignmentId *string `json:"policy_assignment_id,omitempty"`
	// 评估任务执行状态
	State *string `json:"state,omitempty"`
	// 评估任务开始时间
	StartTime *string `json:"start_time,omitempty"`
	// 评估任务结束时间
	EndTime *string `json:"end_time,omitempty"`
	// 评估任务失败信息
	ErrorMessage   *string `json:"error_message,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ShowEvaluationStateByAssignmentIdResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowEvaluationStateByAssignmentIdResponse struct{}"
	}

	return strings.Join([]string{"ShowEvaluationStateByAssignmentIdResponse", string(data)}, " ")
}

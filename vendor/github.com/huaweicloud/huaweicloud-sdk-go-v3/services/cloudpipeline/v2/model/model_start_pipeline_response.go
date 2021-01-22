/*
 * CloudPipeline
 *
 * devcloud pipeline api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type StartPipelineResponse struct {
	// 执行ID
	BuildId *string `json:"build_id,omitempty"`
	// 流水线ID
	PipelineId *string `json:"pipeline_id,omitempty"`
	// 执行时间
	CreateAt *string `json:"create_at,omitempty"`
	// 八爪鱼JobId
	JobId *string `json:"job_id,omitempty"`
	// 八爪鱼JobName
	JobName *string `json:"job_name,omitempty"`
	// 执行人ID
	ExecutorId *string `json:"executor_id,omitempty"`
	// 执行人
	Executor *string `json:"executor,omitempty"`
	// 流水线名字
	PipelineName   *string `json:"pipeline_name,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o StartPipelineResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "StartPipelineResponse struct{}"
	}

	return strings.Join([]string{"StartPipelineResponse", string(data)}, " ")
}

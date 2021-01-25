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

type PipelineBuildResult struct {
	// 执行ID
	BuildId string `json:"build_id"`
	// 运行耗时
	ElapseTime *string `json:"elapse_time,omitempty"`
	// 执行结束时间
	EndTime string `json:"end_time"`
	// 运行结果
	Outcome string `json:"outcome"`
	// 流水线id
	PipelineId string `json:"pipeline_id"`
}

func (o PipelineBuildResult) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PipelineBuildResult struct{}"
	}

	return strings.Join([]string{"PipelineBuildResult", string(data)}, " ")
}

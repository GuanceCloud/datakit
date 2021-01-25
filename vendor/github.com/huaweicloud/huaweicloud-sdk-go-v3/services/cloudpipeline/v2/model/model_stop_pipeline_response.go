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
type StopPipelineResponse struct {
	// 流水线ID
	PipelineId *string `json:"pipeline_id,omitempty"`
	// 流水线名字
	PipelineName   *string `json:"pipeline_name,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o StopPipelineResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "StopPipelineResponse struct{}"
	}

	return strings.Join([]string{"StopPipelineResponse", string(data)}, " ")
}

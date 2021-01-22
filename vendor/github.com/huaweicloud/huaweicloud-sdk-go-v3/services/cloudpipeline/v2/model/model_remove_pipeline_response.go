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
type RemovePipelineResponse struct {
	// 流水线ID
	PipelineId *string `json:"pipeline_id,omitempty"`
	// 流水线名字
	PipelineName   *string `json:"pipeline_name,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o RemovePipelineResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RemovePipelineResponse struct{}"
	}

	return strings.Join([]string{"RemovePipelineResponse", string(data)}, " ")
}

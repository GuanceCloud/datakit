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

// Request Object
type RemovePipelineRequest struct {
	XLanguage  *string `json:"X-Language,omitempty"`
	PipelineId string  `json:"pipeline_id"`
}

func (o RemovePipelineRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RemovePipelineRequest struct{}"
	}

	return strings.Join([]string{"RemovePipelineRequest", string(data)}, " ")
}

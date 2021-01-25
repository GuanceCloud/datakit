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
type StartPipelineRequest struct {
	XLanguage  *string                  `json:"X-Language,omitempty"`
	PipelineId string                   `json:"pipeline_id"`
	Body       *StartPipelineParameters `json:"body,omitempty"`
}

func (o StartPipelineRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "StartPipelineRequest struct{}"
	}

	return strings.Join([]string{"StartPipelineRequest", string(data)}, " ")
}

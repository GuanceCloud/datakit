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
type StartNewPipelineRequest struct {
	XLanguage  *string                  `json:"X-Language,omitempty"`
	PipelineId string                   `json:"pipeline_id"`
	Body       *StartPipelineParameters `json:"body,omitempty"`
}

func (o StartNewPipelineRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "StartNewPipelineRequest struct{}"
	}

	return strings.Join([]string{"StartNewPipelineRequest", string(data)}, " ")
}

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
type BatchShowPipelinesStatusRequest struct {
	XLanguage   *string `json:"X-Language,omitempty"`
	PipelineIds string  `json:"pipeline_ids"`
}

func (o BatchShowPipelinesStatusRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchShowPipelinesStatusRequest struct{}"
	}

	return strings.Join([]string{"BatchShowPipelinesStatusRequest", string(data)}, " ")
}

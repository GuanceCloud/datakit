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
type ShowPipleineStatusRequest struct {
	XLanguage  *string `json:"X-Language,omitempty"`
	PipelineId string  `json:"pipeline_id"`
	BuildId    *string `json:"build_id,omitempty"`
}

func (o ShowPipleineStatusRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowPipleineStatusRequest struct{}"
	}

	return strings.Join([]string{"ShowPipleineStatusRequest", string(data)}, " ")
}

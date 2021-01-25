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
type BatchShowPipelinesStatusResponse struct {
	Body           *[]PipelineExecuteStates `json:"body,omitempty"`
	HttpStatusCode int                      `json:"-"`
}

func (o BatchShowPipelinesStatusResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchShowPipelinesStatusResponse struct{}"
	}

	return strings.Join([]string{"BatchShowPipelinesStatusResponse", string(data)}, " ")
}

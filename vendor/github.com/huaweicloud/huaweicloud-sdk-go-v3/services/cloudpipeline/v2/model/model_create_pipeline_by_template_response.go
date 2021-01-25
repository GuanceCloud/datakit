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
type CreatePipelineByTemplateResponse struct {
	// 实例ID
	TaskId         *string `json:"task_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o CreatePipelineByTemplateResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreatePipelineByTemplateResponse struct{}"
	}

	return strings.Join([]string{"CreatePipelineByTemplateResponse", string(data)}, " ")
}

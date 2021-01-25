/*
 * DevStar
 *
 * DevStar API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type RunCodehubTemplateJobResponse struct {
	// 任务id
	JobId          *string `json:"job_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o RunCodehubTemplateJobResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RunCodehubTemplateJobResponse struct{}"
	}

	return strings.Join([]string{"RunCodehubTemplateJobResponse", string(data)}, " ")
}

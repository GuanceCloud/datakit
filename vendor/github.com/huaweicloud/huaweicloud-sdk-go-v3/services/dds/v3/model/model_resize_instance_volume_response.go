/*
 * DDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ResizeInstanceVolumeResponse struct {
	// 工作流ID。
	JobId          *string `json:"job_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ResizeInstanceVolumeResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResizeInstanceVolumeResponse struct{}"
	}

	return strings.Join([]string{"ResizeInstanceVolumeResponse", string(data)}, " ")
}

/*
 * RDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ListJobInfoDetailRequest struct {
	XLanguage  *string `json:"X-Language,omitempty"`
	InstanceId string  `json:"instance_id"`
	StartTime  string  `json:"start_time"`
	EndTime    *string `json:"end_time,omitempty"`
}

func (o ListJobInfoDetailRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListJobInfoDetailRequest struct{}"
	}

	return strings.Join([]string{"ListJobInfoDetailRequest", string(data)}, " ")
}

/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ListBackgroundTaskRequest struct {
	InstanceId string  `json:"instance_id"`
	Offset     *int32  `json:"offset,omitempty"`
	Limit      *int32  `json:"limit,omitempty"`
	StartTime  *string `json:"start_time,omitempty"`
	EndTime    *string `json:"end_time,omitempty"`
}

func (o ListBackgroundTaskRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListBackgroundTaskRequest struct{}"
	}

	return strings.Join([]string{"ListBackgroundTaskRequest", string(data)}, " ")
}

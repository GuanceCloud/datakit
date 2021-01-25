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
type AddShardingNodeResponse struct {
	// 任务ID。
	JobId          *string `json:"job_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o AddShardingNodeResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AddShardingNodeResponse struct{}"
	}

	return strings.Join([]string{"AddShardingNodeResponse", string(data)}, " ")
}

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

// Response Object
type StartInstanceActionResponse struct {
	// 任务ID。
	JobId          *string `json:"job_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o StartInstanceActionResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "StartInstanceActionResponse struct{}"
	}

	return strings.Join([]string{"StartInstanceActionResponse", string(data)}, " ")
}

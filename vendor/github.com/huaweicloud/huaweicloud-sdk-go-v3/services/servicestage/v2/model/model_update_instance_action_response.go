/*
 * ServiceStage
 *
 * ServiceStage的API,包括应用管理和仓库授权管理
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type UpdateInstanceActionResponse struct {
	// Job ID。
	JobId          *string `json:"job_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o UpdateInstanceActionResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateInstanceActionResponse struct{}"
	}

	return strings.Join([]string{"UpdateInstanceActionResponse", string(data)}, " ")
}

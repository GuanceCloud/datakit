/*
 * SMN
 *
 * SMN Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type UpdateTopicResponse struct {
	// 请求的唯一标识ID。
	RequestId      *string `json:"request_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o UpdateTopicResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateTopicResponse struct{}"
	}

	return strings.Join([]string{"UpdateTopicResponse", string(data)}, " ")
}

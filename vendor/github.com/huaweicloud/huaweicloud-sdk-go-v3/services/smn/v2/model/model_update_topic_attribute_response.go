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
type UpdateTopicAttributeResponse struct {
	// 请求的唯一标识ID。
	RequestId      *string `json:"request_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o UpdateTopicAttributeResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateTopicAttributeResponse struct{}"
	}

	return strings.Join([]string{"UpdateTopicAttributeResponse", string(data)}, " ")
}

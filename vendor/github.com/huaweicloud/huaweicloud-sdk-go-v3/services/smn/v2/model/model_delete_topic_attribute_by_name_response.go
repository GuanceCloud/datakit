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
type DeleteTopicAttributeByNameResponse struct {
	// 请求的唯一标识ID。
	RequestId      *string `json:"request_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o DeleteTopicAttributeByNameResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteTopicAttributeByNameResponse struct{}"
	}

	return strings.Join([]string{"DeleteTopicAttributeByNameResponse", string(data)}, " ")
}

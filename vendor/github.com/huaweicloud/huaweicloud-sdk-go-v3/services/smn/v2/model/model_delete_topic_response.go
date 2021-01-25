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
type DeleteTopicResponse struct {
	// 请求的唯一标识ID
	RequestId      *string `json:"request_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o DeleteTopicResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteTopicResponse struct{}"
	}

	return strings.Join([]string{"DeleteTopicResponse", string(data)}, " ")
}

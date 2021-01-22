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
type PublishMessageResponse struct {
	// 请求的唯一标识ID。
	RequestId *string `json:"request_id,omitempty"`
	// 唯一的消息ID。
	MessageId      *string `json:"message_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o PublishMessageResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PublishMessageResponse struct{}"
	}

	return strings.Join([]string{"PublishMessageResponse", string(data)}, " ")
}

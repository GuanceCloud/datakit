/*
 * DMS
 *
 * DMS Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type SendMessagesResponse struct {
	// 消息列表。
	Messages       *[]SendMessagesRespMessages `json:"messages,omitempty"`
	HttpStatusCode int                         `json:"-"`
}

func (o SendMessagesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SendMessagesResponse struct{}"
	}

	return strings.Join([]string{"SendMessagesResponse", string(data)}, " ")
}

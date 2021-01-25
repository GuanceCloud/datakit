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
type ConsumeMessagesResponse struct {
	// 消息数组。
	Body           *[]ConsumeMessage `json:"body,omitempty"`
	HttpStatusCode int               `json:"-"`
}

func (o ConsumeMessagesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ConsumeMessagesResponse struct{}"
	}

	return strings.Join([]string{"ConsumeMessagesResponse", string(data)}, " ")
}

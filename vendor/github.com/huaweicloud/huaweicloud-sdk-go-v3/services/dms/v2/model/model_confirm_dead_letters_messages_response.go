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
type ConfirmDeadLettersMessagesResponse struct {
	// 确认成功的数目（如果为N，则表示前N条消息确认成功）。
	Success *int32 `json:"success,omitempty"`
	// 确认失败的数目（如果为N，则表示后N条消息确认失败）。
	Fail           *int32 `json:"fail,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ConfirmDeadLettersMessagesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ConfirmDeadLettersMessagesResponse struct{}"
	}

	return strings.Join([]string{"ConfirmDeadLettersMessagesResponse", string(data)}, " ")
}

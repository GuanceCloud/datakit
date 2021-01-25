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

type ConsumeMessage struct {
	Message *ConsumeMessageMessage `json:"message,omitempty"`
	// 消息handler。
	Handler *string `json:"handler,omitempty"`
}

func (o ConsumeMessage) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ConsumeMessage struct{}"
	}

	return strings.Join([]string{"ConsumeMessage", string(data)}, " ")
}

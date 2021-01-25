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
type ConsumeDeadlettersMessageResponse struct {
	// 消息数组。
	Body           *[]ConsumeDeadlettersMessage `json:"body,omitempty"`
	HttpStatusCode int                          `json:"-"`
}

func (o ConsumeDeadlettersMessageResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ConsumeDeadlettersMessageResponse struct{}"
	}

	return strings.Join([]string{"ConsumeDeadlettersMessageResponse", string(data)}, " ")
}

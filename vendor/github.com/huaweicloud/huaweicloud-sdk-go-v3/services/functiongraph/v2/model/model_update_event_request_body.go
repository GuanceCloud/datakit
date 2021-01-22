/*
 * FunctionGraph
 *
 * API v2
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 更新测试事件请求体。
type UpdateEventRequestBody struct {
	// 测试事件content。
	Content *string `json:"content,omitempty"`
}

func (o UpdateEventRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateEventRequestBody struct{}"
	}

	return strings.Join([]string{"UpdateEventRequestBody", string(data)}, " ")
}

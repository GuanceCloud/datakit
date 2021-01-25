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

// Response Object
type ShowEventResponse struct {
	// 测试事件ID。
	Id *string `json:"id,omitempty"`
	// 测试事件名称。
	Name *string `json:"name,omitempty"`
	// 测试事件content。
	Content *string `json:"content,omitempty"`
	// 上次修改的时间。
	LastModified   float32 `json:"last_modified,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ShowEventResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowEventResponse struct{}"
	}

	return strings.Join([]string{"ShowEventResponse", string(data)}, " ")
}

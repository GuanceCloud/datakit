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

// 创建测试事件请求体。
type CreateEventRequestBody struct {
	// 测试事件名称。只能由字母、数字、中划线和下划线组成，且必须以大写或小写字母开头。
	Name *string `json:"name,omitempty"`
	// 测试事件content。
	Content *string `json:"content,omitempty"`
}

func (o CreateEventRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateEventRequestBody struct{}"
	}

	return strings.Join([]string{"CreateEventRequestBody", string(data)}, " ")
}

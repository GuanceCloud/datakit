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
type ListEventsResponse struct {
	// 测试事件总数。
	Count *int32 `json:"count,omitempty"`
	// 测试事件列表。
	Events *[]ListEventsResult `json:"events,omitempty"`
	// 下次读取位置。
	NextMarker     *int64 `json:"next_marker,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListEventsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListEventsResponse struct{}"
	}

	return strings.Join([]string{"ListEventsResponse", string(data)}, " ")
}

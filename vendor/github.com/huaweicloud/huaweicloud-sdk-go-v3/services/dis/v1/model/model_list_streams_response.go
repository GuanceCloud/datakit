/*
 * DIS
 *
 * DIS v1 API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListStreamsResponse struct {
	// 当前租户所有通道数量。
	TotalNumber *int64 `json:"total_number,omitempty"`
	// 满足当前请求条件的通道名称的列表。
	StreamNames *[]string `json:"stream_names,omitempty"`
	// 是否还有更多满足条件的通道。  - true：是 - false：否
	HasMoreStreams *bool `json:"has_more_streams,omitempty"`
	// 通道列表详情。
	StreamInfoList *[]StreamInfo `json:"stream_info_list,omitempty"`
	HttpStatusCode int           `json:"-"`
}

func (o ListStreamsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListStreamsResponse struct{}"
	}

	return strings.Join([]string{"ListStreamsResponse", string(data)}, " ")
}

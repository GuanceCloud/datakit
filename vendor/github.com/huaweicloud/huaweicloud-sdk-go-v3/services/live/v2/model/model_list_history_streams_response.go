/*
 * Live
 *
 * 数据分析服务接口
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListHistoryStreamsResponse struct {
	// 历史流信息列表。
	HistoryStreamList *[]HistoryStreamInfo `json:"history_stream_list,omitempty"`
	// 总记录数
	Total          *int32  `json:"total,omitempty"`
	XRequestId     *string `json:"X-request-id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ListHistoryStreamsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListHistoryStreamsResponse struct{}"
	}

	return strings.Join([]string{"ListHistoryStreamsResponse", string(data)}, " ")
}

/*
 * Live
 *
 * 直播服务源站所有接口
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListLiveStreamsOnlineResponse struct {
	// 总条页数
	TotalPage *int64 `json:"total_page,omitempty"`
	// 总条目数
	TotalNum *int64 `json:"total_num,omitempty"`
	// 偏移量
	Offset *int64 `json:"offset,omitempty"`
	// 每页条目数
	Limit *int64 `json:"limit,omitempty"`
	// 请求唯一标识
	RequestId      *string     `json:"request_id,omitempty"`
	Streams        *OnlineInfo `json:"streams,omitempty"`
	HttpStatusCode int         `json:"-"`
}

func (o ListLiveStreamsOnlineResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListLiveStreamsOnlineResponse struct{}"
	}

	return strings.Join([]string{"ListLiveStreamsOnlineResponse", string(data)}, " ")
}

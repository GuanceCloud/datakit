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

// Request Object
type ListLiveStreamsOnlineRequest struct {
	PublishDomain string  `json:"publish_domain"`
	App           *string `json:"app,omitempty"`
	Offset        *int32  `json:"offset,omitempty"`
	Limit         *int32  `json:"limit,omitempty"`
	Stream        *string `json:"stream,omitempty"`
}

func (o ListLiveStreamsOnlineRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListLiveStreamsOnlineRequest struct{}"
	}

	return strings.Join([]string{"ListLiveStreamsOnlineRequest", string(data)}, " ")
}

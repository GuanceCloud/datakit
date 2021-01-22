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
type ShowOnlineUsersResponse struct {
	// 查询结果的总元素数量
	Total *int32 `json:"total,omitempty"`
	// 正在推流的音视频信息
	UserInfo       *[]UserInfo `json:"user_info,omitempty"`
	HttpStatusCode int         `json:"-"`
}

func (o ShowOnlineUsersResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowOnlineUsersResponse struct{}"
	}

	return strings.Join([]string{"ShowOnlineUsersResponse", string(data)}, " ")
}

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

type UserInfo struct {
	// 直播流的在线人数
	UserNum int32 `json:"user_num"`
	// 操作执行的时间，UTC时间，格式：yyyy-MM-ddTHH:mm:ssZ
	Timestamp string `json:"timestamp"`
}

func (o UserInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UserInfo struct{}"
	}

	return strings.Join([]string{"UserInfo", string(data)}, " ")
}

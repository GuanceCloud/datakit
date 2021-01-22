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
type ShowStreamPortraitResponse struct {
	// 播放画像信息列表。
	StreamPortraits *[]StreamPortrait `json:"stream_portraits,omitempty"`
	XRequestId      *string           `json:"X-request-id,omitempty"`
	HttpStatusCode  int               `json:"-"`
}

func (o ShowStreamPortraitResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowStreamPortraitResponse struct{}"
	}

	return strings.Join([]string{"ShowStreamPortraitResponse", string(data)}, " ")
}

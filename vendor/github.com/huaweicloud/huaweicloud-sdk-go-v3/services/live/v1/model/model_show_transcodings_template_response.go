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
type ShowTranscodingsTemplateResponse struct {
	// 查询结果的总元素数量
	Total *int32 `json:"total,omitempty"`
	// 播放域名
	Domain *string `json:"domain,omitempty"`
	// 转码模板
	Templates      *[]AppQualityInfo `json:"templates,omitempty"`
	HttpStatusCode int               `json:"-"`
}

func (o ShowTranscodingsTemplateResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowTranscodingsTemplateResponse struct{}"
	}

	return strings.Join([]string{"ShowTranscodingsTemplateResponse", string(data)}, " ")
}

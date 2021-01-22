/*
 * BMS
 *
 * BMS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// links字段数据结构说明
type LinksInfo struct {
	// 快捷链接标记名称。取值为：self：包含版本号的资源链接，需要立即跟踪时使用此类链接。bookmark：提供了适合长期存储的资源链接。
	Rel *string `json:"rel,omitempty"`
	// 对应快捷链接
	Href *string `json:"href,omitempty"`
	// 快捷链接类型
	Type *string `json:"type,omitempty"`
}

func (o LinksInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "LinksInfo struct{}"
	}

	return strings.Join([]string{"LinksInfo", string(data)}, " ")
}

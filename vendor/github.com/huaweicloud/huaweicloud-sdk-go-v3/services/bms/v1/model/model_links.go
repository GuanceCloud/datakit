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
type Links struct {
	// 快捷链接标记名称
	Rel *string `json:"rel,omitempty"`
	// 对应快捷链接
	Href *string `json:"href,omitempty"`
}

func (o Links) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Links struct{}"
	}

	return strings.Join([]string{"Links", string(data)}, " ")
}

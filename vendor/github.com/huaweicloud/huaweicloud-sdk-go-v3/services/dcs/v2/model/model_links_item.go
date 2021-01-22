/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type LinksItem struct {
	// 备份文件名称。
	FileName *string `json:"file_name,omitempty"`
	// 备份文件下载链接地址。
	Link *string `json:"link,omitempty"`
}

func (o LinksItem) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "LinksItem struct{}"
	}

	return strings.Join([]string{"LinksItem", string(data)}, " ")
}

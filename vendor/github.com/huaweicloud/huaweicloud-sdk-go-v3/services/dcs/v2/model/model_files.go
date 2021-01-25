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

// 导入的备份文件文件列表。
type Files struct {
	// 备份文件名。
	FileName string `json:"file_name"`
	// 文件大小（单位：Byte）。
	Size *string `json:"size,omitempty"`
	// 文件最后修改时间（格式YYYY-MM-DD HH:MM:SS）。
	UpdateAt *string `json:"update_at,omitempty"`
}

func (o Files) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Files struct{}"
	}

	return strings.Join([]string{"Files", string(data)}, " ")
}

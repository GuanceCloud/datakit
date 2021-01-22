/*
 * DevStar
 *
 * DevStar API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ShowTemplateFileResponse struct {
	// 文件内容
	Content *string `json:"content,omitempty"`
	// 内容编码格式(固定base64)
	Encoding *string `json:"encoding,omitempty"`
	// 文件名
	FileName *string `json:"file_name,omitempty"`
	// 文件相对路径
	FilePath *string `json:"file_path,omitempty"`
	// 文件类型
	FileType       *string `json:"file_type,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ShowTemplateFileResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowTemplateFileResponse struct{}"
	}

	return strings.Join([]string{"ShowTemplateFileResponse", string(data)}, " ")
}

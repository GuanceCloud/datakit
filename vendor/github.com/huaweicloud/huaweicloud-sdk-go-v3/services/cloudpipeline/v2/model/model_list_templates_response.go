/*
 * CloudPipeline
 *
 * devcloud pipeline api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListTemplatesResponse struct {
	// 总数
	Total *int32 `json:"total,omitempty"`
	// 页码数
	PageNumber *int32 `json:"page_number,omitempty"`
	// 每页显示数
	PageSize *int32 `json:"page_size,omitempty"`
	// 模板数据,list类型数据
	Content        *[]TemplateView `json:"content,omitempty"`
	HttpStatusCode int             `json:"-"`
}

func (o ListTemplatesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListTemplatesResponse struct{}"
	}

	return strings.Join([]string{"ListTemplatesResponse", string(data)}, " ")
}

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
type ListTemplatesV2Response struct {
	// 返回模板的数量
	Count *int32 `json:"count,omitempty"`
	// 返回模板的列表
	Templates      *[]TemplateInfo `json:"templates,omitempty"`
	HttpStatusCode int             `json:"-"`
}

func (o ListTemplatesV2Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListTemplatesV2Response struct{}"
	}

	return strings.Join([]string{"ListTemplatesV2Response", string(data)}, " ")
}

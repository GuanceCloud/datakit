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
type ListTemplateViewHistoriesResponse struct {
	// 我浏览的模板
	Templates *[]TemplateViewHistory `json:"templates,omitempty"`
	// 我浏览的模板数量
	Count          *int32 `json:"count,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListTemplateViewHistoriesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListTemplateViewHistoriesResponse struct{}"
	}

	return strings.Join([]string{"ListTemplateViewHistoriesResponse", string(data)}, " ")
}

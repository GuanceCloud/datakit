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
type CreateTemplateViewHistoriesResponse struct {
	// 我浏览的模板
	Templates *[]TemplateViewHistory `json:"templates,omitempty"`
	// 我浏览的模板数量
	Count          *int32 `json:"count,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o CreateTemplateViewHistoriesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateTemplateViewHistoriesResponse struct{}"
	}

	return strings.Join([]string{"CreateTemplateViewHistoriesResponse", string(data)}, " ")
}

/*
 * SMN
 *
 * SMN Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListMessageTemplatesResponse struct {
	// 请求的唯一标识ID。
	RequestId *string `json:"request_id,omitempty"`
	// 返回的模板个数。
	MessageTemplateCount *int32 `json:"message_template_count,omitempty"`
	// Message_template结构体数组。
	MessageTemplates *[]MessageTemplate `json:"message_templates,omitempty"`
	HttpStatusCode   int                `json:"-"`
}

func (o ListMessageTemplatesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListMessageTemplatesResponse struct{}"
	}

	return strings.Join([]string{"ListMessageTemplatesResponse", string(data)}, " ")
}

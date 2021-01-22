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
type CreateMessageTemplateResponse struct {
	// 请求的唯一标识ID。
	RequestId *string `json:"request_id,omitempty"`
	// 模板唯一的资源标识。
	MessageTemplateId *string `json:"message_template_id,omitempty"`
	HttpStatusCode    int     `json:"-"`
}

func (o CreateMessageTemplateResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateMessageTemplateResponse struct{}"
	}

	return strings.Join([]string{"CreateMessageTemplateResponse", string(data)}, " ")
}

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

// Request Object
type DeleteMessageTemplateRequest struct {
	MessageTemplateId string `json:"message_template_id"`
}

func (o DeleteMessageTemplateRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteMessageTemplateRequest struct{}"
	}

	return strings.Join([]string{"DeleteMessageTemplateRequest", string(data)}, " ")
}

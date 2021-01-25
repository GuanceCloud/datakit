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
type CreateMessageTemplateRequest struct {
	Body *CreateMessageTemplateRequestBody `json:"body,omitempty"`
}

func (o CreateMessageTemplateRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateMessageTemplateRequest struct{}"
	}

	return strings.Join([]string{"CreateMessageTemplateRequest", string(data)}, " ")
}

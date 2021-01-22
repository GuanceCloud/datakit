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
type ListMessageTemplatesRequest struct {
	Offset              *int32  `json:"offset,omitempty"`
	Limit               *int32  `json:"limit,omitempty"`
	MessageTemplateName *string `json:"message_template_name,omitempty"`
	Protocol            string  `json:"protocol"`
}

func (o ListMessageTemplatesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListMessageTemplatesRequest struct{}"
	}

	return strings.Join([]string{"ListMessageTemplatesRequest", string(data)}, " ")
}

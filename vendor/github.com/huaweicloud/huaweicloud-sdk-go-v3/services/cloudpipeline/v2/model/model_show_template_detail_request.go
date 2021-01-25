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

// Request Object
type ShowTemplateDetailRequest struct {
	XLanguage    *string `json:"X-Language,omitempty"`
	TemplateId   string  `json:"template_id"`
	TemplateType string  `json:"template_type"`
	Source       *string `json:"source,omitempty"`
}

func (o ShowTemplateDetailRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowTemplateDetailRequest struct{}"
	}

	return strings.Join([]string{"ShowTemplateDetailRequest", string(data)}, " ")
}

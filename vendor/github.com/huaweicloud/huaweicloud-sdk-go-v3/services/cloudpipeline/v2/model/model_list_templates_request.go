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
type ListTemplatesRequest struct {
	XLanguage    *string `json:"X-Language,omitempty"`
	TemplateType string  `json:"template_type"`
	IsBuildIn    string  `json:"is_build_in"`
	Offset       *int32  `json:"offset,omitempty"`
	Limit        *int32  `json:"limit,omitempty"`
	Name         *string `json:"name,omitempty"`
	Sort         *string `json:"sort,omitempty"`
	Asc          *string `json:"asc,omitempty"`
}

func (o ListTemplatesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListTemplatesRequest struct{}"
	}

	return strings.Join([]string{"ListTemplatesRequest", string(data)}, " ")
}

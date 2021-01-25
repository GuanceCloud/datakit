/*
 * BSS
 *
 * Business Support System API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ListServiceTypesRequest struct {
	XLanguage       *string `json:"X-Language,omitempty"`
	ServiceTypeCode *string `json:"service_type_code,omitempty"`
}

func (o ListServiceTypesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListServiceTypesRequest struct{}"
	}

	return strings.Join([]string{"ListServiceTypesRequest", string(data)}, " ")
}

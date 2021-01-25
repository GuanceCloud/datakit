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
type ListResourceTypesRequest struct {
	XLanguage        *string `json:"X-Language,omitempty"`
	ResourceTypeCode *string `json:"resource_type_code,omitempty"`
}

func (o ListResourceTypesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListResourceTypesRequest struct{}"
	}

	return strings.Join([]string{"ListResourceTypesRequest", string(data)}, " ")
}

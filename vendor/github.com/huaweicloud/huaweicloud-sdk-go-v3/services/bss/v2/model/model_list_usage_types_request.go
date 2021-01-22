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
type ListUsageTypesRequest struct {
	XLanguage        *string `json:"X-Language,omitempty"`
	ResourceTypeCode *string `json:"resource_type_code,omitempty"`
	Offset           *int32  `json:"offset,omitempty"`
	Limit            *int32  `json:"limit,omitempty"`
}

func (o ListUsageTypesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListUsageTypesRequest struct{}"
	}

	return strings.Join([]string{"ListUsageTypesRequest", string(data)}, " ")
}

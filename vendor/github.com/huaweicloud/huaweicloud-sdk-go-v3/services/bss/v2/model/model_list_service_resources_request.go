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
type ListServiceResourcesRequest struct {
	XLanguage       *string `json:"X-Language,omitempty"`
	ServiceTypeCode string  `json:"service_type_code"`
	Limit           *int32  `json:"limit,omitempty"`
	Offset          *int32  `json:"offset,omitempty"`
}

func (o ListServiceResourcesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListServiceResourcesRequest struct{}"
	}

	return strings.Join([]string{"ListServiceResourcesRequest", string(data)}, " ")
}

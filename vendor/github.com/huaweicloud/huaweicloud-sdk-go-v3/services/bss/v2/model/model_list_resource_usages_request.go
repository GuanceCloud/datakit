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
type ListResourceUsagesRequest struct {
	XLanguage *string `json:"X-Language,omitempty"`
}

func (o ListResourceUsagesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListResourceUsagesRequest struct{}"
	}

	return strings.Join([]string{"ListResourceUsagesRequest", string(data)}, " ")
}

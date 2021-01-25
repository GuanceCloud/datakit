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
type UpdatePostalRequest struct {
	XLanguage *string          `json:"X-Language,omitempty"`
	Body      *UpdatePostalReq `json:"body,omitempty"`
}

func (o UpdatePostalRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdatePostalRequest struct{}"
	}

	return strings.Join([]string{"UpdatePostalRequest", string(data)}, " ")
}

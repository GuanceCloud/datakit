/*
 * RDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type RestoreToExistingInstanceRequest struct {
	XLanguage *string                               `json:"X-Language,omitempty"`
	Body      *RestoreToExistingInstanceRequestBody `json:"body,omitempty"`
}

func (o RestoreToExistingInstanceRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RestoreToExistingInstanceRequest struct{}"
	}

	return strings.Join([]string{"RestoreToExistingInstanceRequest", string(data)}, " ")
}

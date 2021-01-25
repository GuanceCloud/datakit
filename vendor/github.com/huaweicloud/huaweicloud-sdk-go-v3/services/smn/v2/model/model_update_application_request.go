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
type UpdateApplicationRequest struct {
	ApplicationUrn string                        `json:"application_urn"`
	Body           *UpdateApplicationRequestBody `json:"body,omitempty"`
}

func (o UpdateApplicationRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateApplicationRequest struct{}"
	}

	return strings.Join([]string{"UpdateApplicationRequest", string(data)}, " ")
}

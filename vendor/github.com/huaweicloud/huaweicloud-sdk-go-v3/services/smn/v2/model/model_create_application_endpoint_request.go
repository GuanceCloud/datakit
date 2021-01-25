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
type CreateApplicationEndpointRequest struct {
	ApplicationUrn string                                `json:"application_urn"`
	Body           *CreateApplicationEndpointRequestBody `json:"body,omitempty"`
}

func (o CreateApplicationEndpointRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateApplicationEndpointRequest struct{}"
	}

	return strings.Join([]string{"CreateApplicationEndpointRequest", string(data)}, " ")
}

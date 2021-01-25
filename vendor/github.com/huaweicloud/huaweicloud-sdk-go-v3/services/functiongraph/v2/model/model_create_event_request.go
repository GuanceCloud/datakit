/*
 * FunctionGraph
 *
 * API v2
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type CreateEventRequest struct {
	FunctionUrn string                  `json:"function_urn"`
	Body        *CreateEventRequestBody `json:"body,omitempty"`
}

func (o CreateEventRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateEventRequest struct{}"
	}

	return strings.Join([]string{"CreateEventRequest", string(data)}, " ")
}

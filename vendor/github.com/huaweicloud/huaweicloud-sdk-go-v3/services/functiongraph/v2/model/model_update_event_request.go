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
type UpdateEventRequest struct {
	EventId     string                  `json:"event_id"`
	FunctionUrn string                  `json:"function_urn"`
	Body        *UpdateEventRequestBody `json:"body,omitempty"`
}

func (o UpdateEventRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateEventRequest struct{}"
	}

	return strings.Join([]string{"UpdateEventRequest", string(data)}, " ")
}

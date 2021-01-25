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
type ShowEventRequest struct {
	EventId     string `json:"event_id"`
	FunctionUrn string `json:"function_urn"`
}

func (o ShowEventRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowEventRequest struct{}"
	}

	return strings.Join([]string{"ShowEventRequest", string(data)}, " ")
}

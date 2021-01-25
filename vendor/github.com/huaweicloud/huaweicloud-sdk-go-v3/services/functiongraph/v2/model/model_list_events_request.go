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
type ListEventsRequest struct {
	FunctionUrn string `json:"function_urn"`
}

func (o ListEventsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListEventsRequest struct{}"
	}

	return strings.Join([]string{"ListEventsRequest", string(data)}, " ")
}

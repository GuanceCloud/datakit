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

// Response Object
type ListFunctionTriggersResponse struct {
	Body           *[]ListFunctionTriggerResult `json:"body,omitempty"`
	HttpStatusCode int                          `json:"-"`
}

func (o ListFunctionTriggersResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListFunctionTriggersResponse struct{}"
	}

	return strings.Join([]string{"ListFunctionTriggersResponse", string(data)}, " ")
}

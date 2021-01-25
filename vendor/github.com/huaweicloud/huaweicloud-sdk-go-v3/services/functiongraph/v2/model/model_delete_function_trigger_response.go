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
type DeleteFunctionTriggerResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteFunctionTriggerResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteFunctionTriggerResponse struct{}"
	}

	return strings.Join([]string{"DeleteFunctionTriggerResponse", string(data)}, " ")
}

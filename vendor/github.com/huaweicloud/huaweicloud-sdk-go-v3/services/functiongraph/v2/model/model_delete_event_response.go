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
type DeleteEventResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteEventResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteEventResponse struct{}"
	}

	return strings.Join([]string{"DeleteEventResponse", string(data)}, " ")
}

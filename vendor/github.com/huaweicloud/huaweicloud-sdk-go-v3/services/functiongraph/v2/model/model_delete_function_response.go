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
type DeleteFunctionResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteFunctionResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteFunctionResponse struct{}"
	}

	return strings.Join([]string{"DeleteFunctionResponse", string(data)}, " ")
}

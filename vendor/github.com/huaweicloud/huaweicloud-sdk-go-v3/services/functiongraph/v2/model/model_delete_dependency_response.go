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
type DeleteDependencyResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteDependencyResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteDependencyResponse struct{}"
	}

	return strings.Join([]string{"DeleteDependencyResponse", string(data)}, " ")
}

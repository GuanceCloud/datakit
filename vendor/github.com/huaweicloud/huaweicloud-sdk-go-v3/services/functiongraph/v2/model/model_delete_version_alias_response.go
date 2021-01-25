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
type DeleteVersionAliasResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteVersionAliasResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteVersionAliasResponse struct{}"
	}

	return strings.Join([]string{"DeleteVersionAliasResponse", string(data)}, " ")
}

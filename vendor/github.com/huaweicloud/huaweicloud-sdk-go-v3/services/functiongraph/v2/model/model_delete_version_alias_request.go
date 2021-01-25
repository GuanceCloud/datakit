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
type DeleteVersionAliasRequest struct {
	FunctionUrn string `json:"function_urn"`
	Name        string `json:"name"`
}

func (o DeleteVersionAliasRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteVersionAliasRequest struct{}"
	}

	return strings.Join([]string{"DeleteVersionAliasRequest", string(data)}, " ")
}

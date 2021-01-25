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
type ShowVersionAliasRequest struct {
	FunctionUrn string `json:"function_urn"`
	Name        string `json:"name"`
}

func (o ShowVersionAliasRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowVersionAliasRequest struct{}"
	}

	return strings.Join([]string{"ShowVersionAliasRequest", string(data)}, " ")
}

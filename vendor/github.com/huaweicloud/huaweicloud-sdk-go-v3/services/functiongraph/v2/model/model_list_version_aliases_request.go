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
type ListVersionAliasesRequest struct {
	FunctionUrn string `json:"function_urn"`
}

func (o ListVersionAliasesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListVersionAliasesRequest struct{}"
	}

	return strings.Join([]string{"ListVersionAliasesRequest", string(data)}, " ")
}

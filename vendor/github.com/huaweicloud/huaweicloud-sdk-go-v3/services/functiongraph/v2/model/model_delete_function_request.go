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
type DeleteFunctionRequest struct {
	FunctionUrn string `json:"function_urn"`
}

func (o DeleteFunctionRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteFunctionRequest struct{}"
	}

	return strings.Join([]string{"DeleteFunctionRequest", string(data)}, " ")
}

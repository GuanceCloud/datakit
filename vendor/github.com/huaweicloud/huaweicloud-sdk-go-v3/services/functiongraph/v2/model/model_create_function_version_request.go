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
type CreateFunctionVersionRequest struct {
	FunctionUrn string                            `json:"function_urn"`
	Body        *CreateFunctionVersionRequestBody `json:"body,omitempty"`
}

func (o CreateFunctionVersionRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateFunctionVersionRequest struct{}"
	}

	return strings.Join([]string{"CreateFunctionVersionRequest", string(data)}, " ")
}

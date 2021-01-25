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
type UpdateFunctionCodeRequest struct {
	FunctionUrn string                         `json:"function_urn"`
	Body        *UpdateFunctionCodeRequestBody `json:"body,omitempty"`
}

func (o UpdateFunctionCodeRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateFunctionCodeRequest struct{}"
	}

	return strings.Join([]string{"UpdateFunctionCodeRequest", string(data)}, " ")
}

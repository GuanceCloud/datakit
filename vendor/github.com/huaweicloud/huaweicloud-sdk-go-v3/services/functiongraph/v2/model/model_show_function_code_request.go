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
type ShowFunctionCodeRequest struct {
	FunctionUrn string `json:"function_urn"`
}

func (o ShowFunctionCodeRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowFunctionCodeRequest struct{}"
	}

	return strings.Join([]string{"ShowFunctionCodeRequest", string(data)}, " ")
}

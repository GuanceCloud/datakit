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
type ListFunctionTriggersRequest struct {
	FunctionUrn string `json:"function_urn"`
}

func (o ListFunctionTriggersRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListFunctionTriggersRequest struct{}"
	}

	return strings.Join([]string{"ListFunctionTriggersRequest", string(data)}, " ")
}

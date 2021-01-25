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
type BatchDeleteFunctionTriggersRequest struct {
	FunctionUrn string `json:"function_urn"`
}

func (o BatchDeleteFunctionTriggersRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchDeleteFunctionTriggersRequest struct{}"
	}

	return strings.Join([]string{"BatchDeleteFunctionTriggersRequest", string(data)}, " ")
}

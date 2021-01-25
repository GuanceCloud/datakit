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
type UpdateFunctionReservedInstancesRequest struct {
	FunctionUrn string                                      `json:"function_urn"`
	Body        *UpdateFunctionReservedInstancesRequestBody `json:"body,omitempty"`
}

func (o UpdateFunctionReservedInstancesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateFunctionReservedInstancesRequest struct{}"
	}

	return strings.Join([]string{"UpdateFunctionReservedInstancesRequest", string(data)}, " ")
}

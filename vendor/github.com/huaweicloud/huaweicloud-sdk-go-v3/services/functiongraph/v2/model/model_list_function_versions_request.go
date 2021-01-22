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
type ListFunctionVersionsRequest struct {
	FunctionUrn string  `json:"function_urn"`
	Marker      *string `json:"marker,omitempty"`
	Maxitems    *string `json:"maxitems,omitempty"`
}

func (o ListFunctionVersionsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListFunctionVersionsRequest struct{}"
	}

	return strings.Join([]string{"ListFunctionVersionsRequest", string(data)}, " ")
}

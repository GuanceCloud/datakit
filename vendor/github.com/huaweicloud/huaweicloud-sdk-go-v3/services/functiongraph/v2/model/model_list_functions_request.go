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
type ListFunctionsRequest struct {
	Marker   *string `json:"marker,omitempty"`
	Maxitems *string `json:"maxitems,omitempty"`
}

func (o ListFunctionsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListFunctionsRequest struct{}"
	}

	return strings.Join([]string{"ListFunctionsRequest", string(data)}, " ")
}

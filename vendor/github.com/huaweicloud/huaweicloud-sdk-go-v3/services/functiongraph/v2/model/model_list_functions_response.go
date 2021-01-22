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

// Response Object
type ListFunctionsResponse struct {
	// 函数列表。
	Functions *[]ListFunctionResult `json:"functions,omitempty"`
	// 函数下次记录读取位置。
	NextMarker     *int32 `json:"next_marker,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListFunctionsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListFunctionsResponse struct{}"
	}

	return strings.Join([]string{"ListFunctionsResponse", string(data)}, " ")
}

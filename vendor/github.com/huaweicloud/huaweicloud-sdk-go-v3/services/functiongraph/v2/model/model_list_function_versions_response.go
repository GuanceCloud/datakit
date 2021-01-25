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
type ListFunctionVersionsResponse struct {
	// 版本列表
	Versions *[]ListFunctionVersionResult `json:"versions,omitempty"`
	// 下一次记录位置
	NextMarker     *int64 `json:"next_marker,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListFunctionVersionsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListFunctionVersionsResponse struct{}"
	}

	return strings.Join([]string{"ListFunctionVersionsResponse", string(data)}, " ")
}

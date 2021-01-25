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
type ListDependenciesResponse struct {
	// 依赖包总数。
	Count *int32 `json:"count,omitempty"`
	// 依赖包列表。
	Dependencies *[]ListDependenciesResult `json:"dependencies,omitempty"`
	// 下次读取位置。
	NextMarker     *int64 `json:"next_marker,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListDependenciesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListDependenciesResponse struct{}"
	}

	return strings.Join([]string{"ListDependenciesResponse", string(data)}, " ")
}

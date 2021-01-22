/*
 * ServiceStage
 *
 * ServiceStage的API,包括应用管理和仓库授权管理
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListComponentsResponse struct {
	// 组件个数。
	Count *int32 `json:"count,omitempty"`
	// 组件列表。
	Components     *[]ComponentView `json:"components,omitempty"`
	HttpStatusCode int              `json:"-"`
}

func (o ListComponentsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListComponentsResponse struct{}"
	}

	return strings.Join([]string{"ListComponentsResponse", string(data)}, " ")
}

/*
 * BSS
 *
 * Business Support System API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListServiceResourcesResponse struct {
	// |参数名称：总数| |参数的约束及描述：总数|
	TotalCount *int32 `json:"total_count,omitempty"`
	// |参数名称：资源基本信息列表| |参数约束以及描述：资源基本信息列表|
	Infos          *[]ServiceResourceInfo `json:"infos,omitempty"`
	HttpStatusCode int                    `json:"-"`
}

func (o ListServiceResourcesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListServiceResourcesResponse struct{}"
	}

	return strings.Join([]string{"ListServiceResourcesResponse", string(data)}, " ")
}

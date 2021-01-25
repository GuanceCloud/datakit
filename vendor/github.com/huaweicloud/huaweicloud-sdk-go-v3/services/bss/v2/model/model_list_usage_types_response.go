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
type ListUsageTypesResponse struct {
	// |参数名称：总条数，必须大于等于0。| |参数的约束及描述：总条数，必须大于等于0。|
	TotalCount *int32 `json:"total_count,omitempty"`
	// |参数名称：用量类型列表| |参数约束以及描述：用量类型列表|
	UsageTypes     *[]UsageType `json:"usage_types,omitempty"`
	HttpStatusCode int          `json:"-"`
}

func (o ListUsageTypesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListUsageTypesResponse struct{}"
	}

	return strings.Join([]string{"ListUsageTypesResponse", string(data)}, " ")
}

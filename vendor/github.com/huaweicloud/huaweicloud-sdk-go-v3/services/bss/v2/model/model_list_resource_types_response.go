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
type ListResourceTypesResponse struct {
	// |参数名称：返回数据| |参数约束以及描述：返回数据|
	ResourceTypes  *[]ResourceType `json:"resource_types,omitempty"`
	HttpStatusCode int             `json:"-"`
}

func (o ListResourceTypesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListResourceTypesResponse struct{}"
	}

	return strings.Join([]string{"ListResourceTypesResponse", string(data)}, " ")
}

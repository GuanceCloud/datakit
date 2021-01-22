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

type ResourceType struct {
	// |参数名称：资源类型编码| |参数约束及描述：资源类型编码|
	ResourceTypeCode *string `json:"resource_type_code,omitempty"`
	// |参数名称：资源类型名称| |参数约束及描述：资源类型名称|
	ResourceTypeName *string `json:"resource_type_name,omitempty"`
	// |参数名称：资源类型描述| |参数约束及描述：资源类型描述|
	ResourceTypeDesc *string `json:"resource_type_desc,omitempty"`
}

func (o ResourceType) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResourceType struct{}"
	}

	return strings.Join([]string{"ResourceType", string(data)}, " ")
}

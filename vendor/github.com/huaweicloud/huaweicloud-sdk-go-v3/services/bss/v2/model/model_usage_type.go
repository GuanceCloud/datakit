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

type UsageType struct {
	// |参数名称：用量类型编码如：duration| |参数约束及描述：用量类型编码如：duration|
	Code *string `json:"code,omitempty"`
	// |参数名称：用量类型名称| |参数约束及描述：用量类型名称|
	Name *string `json:"name,omitempty"`
	// |参数名称：资源类型编码| |参数约束及描述：资源类型编码|
	ResourceTypeCode *string `json:"resource_type_code,omitempty"`
	// |参数名称：服务类型编码| |参数约束及描述：服务类型编码|
	ServiceTypeCode *string `json:"service_type_code,omitempty"`
}

func (o UsageType) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UsageType struct{}"
	}

	return strings.Join([]string{"UsageType", string(data)}, " ")
}

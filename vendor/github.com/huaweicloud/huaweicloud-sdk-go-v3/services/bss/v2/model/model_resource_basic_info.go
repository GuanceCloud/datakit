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

type ResourceBasicInfo struct {
	// |参数名称：资源类型编码| |参数约束及描述：资源类型编码|
	ResourceTypeCode *string `json:"resource_type_code,omitempty"`
	// |参数名称：资源类型归属的服务类型编码| |参数约束及描述：资源类型归属的服务类型编码|
	ProductOwnerService *string `json:"product_owner_service,omitempty"`
	// |参数名称：资源名称，按照请求的X-Language返回对应语言的名称| |参数约束及描述：资源名称，按照请求的X-Language返回对应语言的名称|
	Name *string `json:"name,omitempty"`
	// |参数名称：资源描述，按照请求的X-Language返回对应语言的描述| |参数约束及描述：资源描述，按照请求的X-Language返回对应语言的描述|
	Description *string `json:"description,omitempty"`
}

func (o ResourceBasicInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResourceBasicInfo struct{}"
	}

	return strings.Join([]string{"ResourceBasicInfo", string(data)}, " ")
}

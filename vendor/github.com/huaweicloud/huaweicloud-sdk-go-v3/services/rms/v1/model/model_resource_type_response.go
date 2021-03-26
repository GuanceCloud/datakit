/*
 * RMS
 *
 * Resource Manager Api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 资源类型详情
type ResourceTypeResponse struct {
	// 资源类型名称
	Name *string `json:"name,omitempty"`
	// 资源类型显示名称，可以通过请求中 'X-Language'设置语言
	DisplayName *string `json:"display_name,omitempty"`
	// 是否是全局类型的资源
	Global *bool `json:"global,omitempty"`
	// 支持的region列表
	Regions *[]string `json:"regions,omitempty"`
	// console终端id
	ConsoleEndpointId *string `json:"console_endpoint_id,omitempty"`
	// console列表页地址
	ConsoleListUrl *string `json:"console_list_url,omitempty"`
	// console详情页地址
	ConsoleDetailUrl *string `json:"console_detail_url,omitempty"`
}

func (o ResourceTypeResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResourceTypeResponse struct{}"
	}

	return strings.Join([]string{"ResourceTypeResponse", string(data)}, " ")
}

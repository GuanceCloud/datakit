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

// 规则过滤器
type PolicyFilterDefinition struct {
	// 区域ID
	RegionId *string `json:"region_id,omitempty"`
	// 资源服务
	ResourceProvider *string `json:"resource_provider,omitempty"`
	// 资源类型
	ResourceType *string `json:"resource_type,omitempty"`
	// 资源ID
	ResourceId *string `json:"resource_id,omitempty"`
	// 标签键
	TagKey *string `json:"tag_key,omitempty"`
	// 标签值
	TagValue *string `json:"tag_value,omitempty"`
}

func (o PolicyFilterDefinition) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PolicyFilterDefinition struct{}"
	}

	return strings.Join([]string{"PolicyFilterDefinition", string(data)}, " ")
}

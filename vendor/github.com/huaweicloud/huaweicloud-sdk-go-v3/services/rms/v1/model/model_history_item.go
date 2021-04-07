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

// 资源历史
type HistoryItem struct {
	// 租户id
	DomainId *string `json:"domain_id,omitempty"`
	// 资源id
	ResourceId *string `json:"resource_id,omitempty"`
	// 资源类型
	ResourceType *string `json:"resource_type,omitempty"`
	// 该资源在RMS系统捕获时间
	CaptureTime *string `json:"capture_time,omitempty"`
	// 资源状态
	Status *string `json:"status,omitempty"`
	// 资源关系列表
	Relations *[]ResourceRelation `json:"relations,omitempty"`
	Resource  *ResourceEntity     `json:"resource,omitempty"`
}

func (o HistoryItem) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "HistoryItem struct{}"
	}

	return strings.Join([]string{"HistoryItem", string(data)}, " ")
}

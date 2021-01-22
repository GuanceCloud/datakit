/*
 * Kafka
 *
 * Kafka Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 子维度信息。
type ShowCeshierarchyRespChildren struct {
	// 子维度名称。
	Name *string `json:"name,omitempty"`
	// 监控指标名称列表。
	Metrics *[]string `json:"metrics,omitempty"`
	// 监控查询使用的key。
	KeyName *[]string `json:"key_name,omitempty"`
	// 监控维度路由。
	DimRouter *[]string `json:"dim_router,omitempty"`
}

func (o ShowCeshierarchyRespChildren) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowCeshierarchyRespChildren struct{}"
	}

	return strings.Join([]string{"ShowCeshierarchyRespChildren", string(data)}, " ")
}

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

type ShowCeshierarchyRespDimensions struct {
	// 监控维度名称。
	Name *string `json:"name,omitempty"`
	// 监控指标名称。请参考[支持的监控指标](https://support.huaweicloud.com/usermanual-kafka/kafka-ug-180413002.html)。
	Metrics *[]string `json:"metrics,omitempty"`
	// 监控查询使用的key。
	KeyName *[]string `json:"key_name,omitempty"`
	// 监控维度路由。
	DimRouter *[]string `json:"dim_router,omitempty"`
	// 子维度列表。
	Children *[]ShowCeshierarchyRespChildren `json:"children,omitempty"`
}

func (o ShowCeshierarchyRespDimensions) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowCeshierarchyRespDimensions struct{}"
	}

	return strings.Join([]string{"ShowCeshierarchyRespDimensions", string(data)}, " ")
}

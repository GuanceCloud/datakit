/*
 * SWR
 *
 * SWR API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 需要创建触发器的信息
type CreateTriggerRequestBody struct {
	// 触发动作，update
	Action string `json:"action"`
	// 应用类型，deployments、statefulsets
	AppType string `json:"app_type"`
	// 应用名
	Application string `json:"application"`
	// 集群ID（trigger_mode设置为cci时为空)
	ClusterId *string `json:"cluster_id,omitempty"`
	// 集群名（trigger_mode设置为cci时为空）
	ClusterName *string `json:"cluster_name,omitempty"`
	// 应用名所在的namespace
	ClusterNs string `json:"cluster_ns"`
	// 触发条件，type为all时为.*,type为tag时为tag名,type为regular时为正则表达式
	Condition string `json:"condition"`
	// 需更新的container名，默认为所有container
	Container *string `json:"container,omitempty"`
	// 是否生效
	Enable string `json:"enable"`
	// 触发器名
	Name string `json:"name"`
	// 触发器类型，cce、cci，默认为cce
	TriggerMode *string `json:"trigger_mode,omitempty"`
	// 触发条件，all、tag、regular
	TriggerType string `json:"trigger_type"`
}

func (o CreateTriggerRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateTriggerRequestBody struct{}"
	}

	return strings.Join([]string{"CreateTriggerRequestBody", string(data)}, " ")
}

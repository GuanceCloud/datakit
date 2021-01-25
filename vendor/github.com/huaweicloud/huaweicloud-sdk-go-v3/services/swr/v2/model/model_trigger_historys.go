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

type TriggerHistorys struct {
	// 触发动作，update
	Action string `json:"action"`
	// 应用类型，deployments、statefulsets
	AppType string `json:"app_type"`
	// 应用名
	Application string `json:"application"`
	// 集群ID（cci时为空）
	ClusterId string `json:"cluster_id"`
	// 集群名（cci时为空）
	ClusterName string `json:"cluster_name"`
	// 应用名所在的namespace
	ClusterNs string `json:"cluster_ns"`
	// 触发条件，type为all时为.*,type为tag时为tag名,type为regular时为正则表达式
	Condition string `json:"condition"`
	// 需更新的container名，默认为所有container
	Container string `json:"container"`
	// 创建时间
	CreatedAt string `json:"created_at"`
	// 创建人
	CreatorName string `json:"creator_name"`
	// 详情
	Detail string `json:"detail"`
	// 更新结果，success、failed
	Result string `json:"result"`
	// 触发的版本号
	Tag string `json:"tag"`
}

func (o TriggerHistorys) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "TriggerHistorys struct{}"
	}

	return strings.Join([]string{"TriggerHistorys", string(data)}, " ")
}

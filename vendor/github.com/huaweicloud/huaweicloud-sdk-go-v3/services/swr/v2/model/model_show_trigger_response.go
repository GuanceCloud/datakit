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

// Response Object
type ShowTriggerResponse struct {
	// 触发动作，update
	Action *string `json:"action,omitempty"`
	// 应用类型，deployments、statefulsets
	AppType *string `json:"app_type,omitempty"`
	// 应用名
	Application *string `json:"application,omitempty"`
	// 集群ID（cci时为空）
	ClusterId *string `json:"cluster_id,omitempty"`
	// 集群名（cci时为空）
	ClusterName *string `json:"cluster_name,omitempty"`
	// 应用名所在的namespace
	ClusterNs *string `json:"cluster_ns,omitempty"`
	// 触发条件，type为all时为.*,type为tag时为tag名,type为regular时为正则表达式
	Condition *string `json:"condition,omitempty"`
	// 需更新的container名，默认为所有container
	Container *string `json:"container,omitempty"`
	// 创建时间
	CreatedAt *string `json:"created_at,omitempty"`
	// 创建人
	CreatorName *string `json:"creator_name,omitempty"`
	// 是否生效
	Enable *string `json:"enable,omitempty"`
	// 触发器名
	Name *string `json:"name,omitempty"`
	// 触发器历史
	TriggerHistory *[]TriggerHistorys `json:"trigger_history,omitempty"`
	// 触发器类型，cce、cci
	TriggerMode *string `json:"trigger_mode,omitempty"`
	// 触发条件，all、tag、regular
	TriggerType    *string `json:"trigger_type,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ShowTriggerResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowTriggerResponse struct{}"
	}

	return strings.Join([]string{"ShowTriggerResponse", string(data)}, " ")
}

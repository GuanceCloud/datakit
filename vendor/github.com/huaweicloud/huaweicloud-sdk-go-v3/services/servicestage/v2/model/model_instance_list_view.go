/*
 * ServiceStage
 *
 * ServiceStage的API,包括应用管理和仓库授权管理
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 实例参数。
type InstanceListView struct {
	// 应用组件实例ID。
	Id *string `json:"id,omitempty"`
	// 应用ID。
	ApplicationId *string `json:"application_id,omitempty"`
	// 应用名称。
	ApplicationName *string `json:"application_name,omitempty"`
	// 组件ID。
	ComponentId *string `json:"component_id,omitempty"`
	// 组件名称。
	ComponentName *string `json:"component_name,omitempty"`
	// 应用组件实例名称。
	Name *string `json:"name,omitempty"`
	// 应用组件环境ID。
	EnvironmentId *string `json:"environment_id,omitempty"`
	// 环境名称。
	EnvironmentName *string `json:"environment_name,omitempty"`
	// 运行平台类型。 应用可以在不同的平台上运行，可选用的平台的类型有以下几种：cce、vmapp。
	PlatformType *string `json:"platform_type,omitempty"`
	// 应用组件版本号。
	Version *string `json:"version,omitempty"`
	// 访问方式。
	ExternalAccesses *[]ExternalAccesses `json:"external_accesses,omitempty"`
	// 组件部署件。key为组件component_name，对于Docker多容器场景，key为容器名称。
	Artifacts map[string]interface{} `json:"artifacts,omitempty"`
	// 创建人。
	Creator *string `json:"creator,omitempty"`
	// 创建时间。
	CreateTime *int64 `json:"create_time,omitempty"`
	// 修改时间。
	UpdateTime   *int64              `json:"update_time,omitempty"`
	StatusDetail *InstanceStatusView `json:"status_detail,omitempty"`
}

func (o InstanceListView) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "InstanceListView struct{}"
	}

	return strings.Join([]string{"InstanceListView", string(data)}, " ")
}

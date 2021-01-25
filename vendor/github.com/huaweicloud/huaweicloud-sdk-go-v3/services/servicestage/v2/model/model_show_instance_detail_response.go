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

// Response Object
type ShowInstanceDetailResponse struct {
	// 应用组件实例ID。
	Id *string `json:"id,omitempty"`
	// 应用组件实例名称。
	Name *string `json:"name,omitempty"`
	// 实例描述。
	Description *string `json:"description,omitempty"`
	// 应用组件环境ID。
	EnvironmentId *string               `json:"environment_id,omitempty"`
	PlatformType  *InstancePlatformType `json:"platform_type,omitempty"`
	FlavorId      *FlavorId             `json:"flavor_id,omitempty"`
	// 组件部署件。key为组件component_name，对于Docker多容器场景，key为容器名称。
	Artifacts map[string]interface{} `json:"artifacts,omitempty"`
	// 应用组件版本号。
	Version *string `json:"version,omitempty"`
	// 应用组件配置，如环境变量。
	Configuration *interface{} `json:"configuration,omitempty"`
	// 创建人。
	Creator *string `json:"creator,omitempty"`
	// 创建时间。
	CreateTime *int64 `json:"create_time,omitempty"`
	// 修改时间。
	UpdateTime *int64 `json:"update_time,omitempty"`
	// 访问方式列表。
	ExternalAccesses *[]ExternalAccesses `json:"external_accesses,omitempty"`
	// 部署资源列表。
	ReferResources *[]ReferResources   `json:"refer_resources,omitempty"`
	StatusDetail   *InstanceStatusView `json:"status_detail,omitempty"`
	HttpStatusCode int                 `json:"-"`
}

func (o ShowInstanceDetailResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowInstanceDetailResponse struct{}"
	}

	return strings.Join([]string{"ShowInstanceDetailResponse", string(data)}, " ")
}

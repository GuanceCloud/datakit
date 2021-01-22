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

type InstanceModify struct {
	// 应用组件版本号，满足版本语义，如1.0.1。
	Version  string    `json:"version"`
	FlavorId *FlavorId `json:"flavor_id,omitempty"`
	// 组件部署件。key为组件component_name，对于Docker多容器场景，key为容器名称。
	Artifacts map[string]interface{} `json:"artifacts,omitempty"`
	// 应用配置，如环境变量。
	Configuration map[string]interface{} `json:"configuration,omitempty"`
	// 描述。
	Description *string `json:"description,omitempty"`
	// 访问方式列表。
	ExternalAccesses *[]ExternalAccesses `json:"external_accesses,omitempty"`
	// 部署资源列表。
	ReferResources *[]ReferResourceCreate `json:"refer_resources,omitempty"`
}

func (o InstanceModify) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "InstanceModify struct{}"
	}

	return strings.Join([]string{"InstanceModify", string(data)}, " ")
}

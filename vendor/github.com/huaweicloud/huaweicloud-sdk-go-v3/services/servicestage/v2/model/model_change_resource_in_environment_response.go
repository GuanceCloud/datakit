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
type ChangeResourceInEnvironmentResponse struct {
	// 环境ID。
	Id *string `json:"id,omitempty"`
	// 环境名称。
	Name *string `json:"name,omitempty"`
	// 环境别名。
	Alias *string `json:"alias,omitempty"`
	// 环境描述。
	Description *string `json:"description,omitempty"`
	// 项目ID。
	ProjectId *string `json:"project_id,omitempty"`
	// 企业项目ID。
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
	// 收费模式。
	ChargeMode *string `json:"charge_mode,omitempty"`
	// 虚拟私有云ID。
	VpcId *string `json:"vpc_id,omitempty"`
	// 基础资源。
	BaseResources *[]Resource `json:"base_resources,omitempty"`
	// 可选资源。
	OptionalResources *[]Resource `json:"optional_resources,omitempty"`
	// 创建人。
	Creator *string `json:"creator,omitempty"`
	// 创建时间。
	CreateTime *int64 `json:"create_time,omitempty"`
	// 修改时间。
	UpdateTime     *int64 `json:"update_time,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ChangeResourceInEnvironmentResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ChangeResourceInEnvironmentResponse struct{}"
	}

	return strings.Join([]string{"ChangeResourceInEnvironmentResponse", string(data)}, " ")
}

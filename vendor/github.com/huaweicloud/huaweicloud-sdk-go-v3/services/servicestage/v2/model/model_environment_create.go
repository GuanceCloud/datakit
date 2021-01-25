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

type EnvironmentCreate struct {
	// 环境名称。
	Name string `json:"name"`
	// 环境别名。
	Alias *string `json:"alias,omitempty"`
	// 环境描述。
	Description *string `json:"description,omitempty"`
	// 企业项目ID。
	EnterpriseProjectId *string     `json:"enterprise_project_id,omitempty"`
	ChargeMode          *ChargeMode `json:"charge_mode,omitempty"`
	// 虚拟私有云ID。
	VpcId string `json:"vpc_id"`
	// 基础资源。
	BaseResources []Resource `json:"base_resources"`
	// 可选资源。
	OptionalResources *[]Resource `json:"optional_resources,omitempty"`
}

func (o EnvironmentCreate) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "EnvironmentCreate struct{}"
	}

	return strings.Join([]string{"EnvironmentCreate", string(data)}, " ")
}

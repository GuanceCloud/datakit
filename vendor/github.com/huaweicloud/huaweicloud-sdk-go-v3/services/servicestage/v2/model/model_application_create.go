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

type ApplicationCreate struct {
	// 应用名称。
	Name string `json:"name"`
	// 应用描述。
	Description *string `json:"description,omitempty"`
	// 企业项目ID。默认值为0。
	EnterpriseProjectId *string `json:"enterprise_project_id,omitempty"`
}

func (o ApplicationCreate) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ApplicationCreate struct{}"
	}

	return strings.Join([]string{"ApplicationCreate", string(data)}, " ")
}

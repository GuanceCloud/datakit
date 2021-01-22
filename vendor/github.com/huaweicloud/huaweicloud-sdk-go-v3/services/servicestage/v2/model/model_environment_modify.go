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

type EnvironmentModify struct {
	// 环境名称。
	Name *string `json:"name,omitempty"`
	// 环境别名。
	Alias *string `json:"alias,omitempty"`
	// 环境描述。
	Description *string `json:"description,omitempty"`
}

func (o EnvironmentModify) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "EnvironmentModify struct{}"
	}

	return strings.Join([]string{"EnvironmentModify", string(data)}, " ")
}

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

type ApplicationModify struct {
	// 应用名称。
	Name *string `json:"name,omitempty"`
	// 应用描述。
	Description *string `json:"description,omitempty"`
}

func (o ApplicationModify) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ApplicationModify struct{}"
	}

	return strings.Join([]string{"ApplicationModify", string(data)}, " ")
}

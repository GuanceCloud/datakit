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

// 应用配置信息。
type ApplicationConfigModifyConfiguration struct {
	// 应用环境变量。
	Env []ApplicationConfigModifyConfigurationEnv `json:"env"`
}

func (o ApplicationConfigModifyConfiguration) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ApplicationConfigModifyConfiguration struct{}"
	}

	return strings.Join([]string{"ApplicationConfigModifyConfiguration", string(data)}, " ")
}

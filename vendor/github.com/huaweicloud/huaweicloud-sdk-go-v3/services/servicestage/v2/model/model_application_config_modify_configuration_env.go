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

type ApplicationConfigModifyConfigurationEnv struct {
	// 环境变量名称。
	Name string `json:"name"`
	// 环境变量取值。
	Value string `json:"value"`
}

func (o ApplicationConfigModifyConfigurationEnv) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ApplicationConfigModifyConfigurationEnv struct{}"
	}

	return strings.Join([]string{"ApplicationConfigModifyConfigurationEnv", string(data)}, " ")
}

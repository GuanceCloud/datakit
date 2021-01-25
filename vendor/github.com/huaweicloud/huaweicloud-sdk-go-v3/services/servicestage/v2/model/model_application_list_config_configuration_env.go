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

type ApplicationListConfigConfigurationEnv struct {
	// 环境变量名称。
	Name *string `json:"name,omitempty"`
	// 环境变量取值。
	Value *string `json:"value,omitempty"`
}

func (o ApplicationListConfigConfigurationEnv) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ApplicationListConfigConfigurationEnv struct{}"
	}

	return strings.Join([]string{"ApplicationListConfigConfigurationEnv", string(data)}, " ")
}

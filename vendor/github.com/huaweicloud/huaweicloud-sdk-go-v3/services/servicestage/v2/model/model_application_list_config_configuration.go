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
type ApplicationListConfigConfiguration struct {
	// 应用组件环境变量。
	Env *[]ApplicationListConfigConfigurationEnv `json:"env,omitempty"`
}

func (o ApplicationListConfigConfiguration) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ApplicationListConfigConfiguration struct{}"
	}

	return strings.Join([]string{"ApplicationListConfigConfiguration", string(data)}, " ")
}

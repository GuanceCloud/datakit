/*
 * APIG
 *
 * API网关（API Gateway）是为开发者、合作伙伴提供的高性能、高可用、高安全的API托管服务，帮助用户轻松构建、管理和发布任意规模的API。
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type CreateEnvironmentVariableV2Response struct {
	// 变量值
	VariableValue *string `json:"variable_value,omitempty"`
	// 环境编号
	EnvId *string `json:"env_id,omitempty"`
	// API分组编号
	GroupId *string `json:"group_id,omitempty"`
	// 环境变量编号
	Id *string `json:"id,omitempty"`
	// 变量名
	VariableName   *string `json:"variable_name,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o CreateEnvironmentVariableV2Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateEnvironmentVariableV2Response struct{}"
	}

	return strings.Join([]string{"CreateEnvironmentVariableV2Response", string(data)}, " ")
}

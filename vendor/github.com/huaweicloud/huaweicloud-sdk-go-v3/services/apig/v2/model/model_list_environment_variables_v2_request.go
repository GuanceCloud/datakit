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

// Request Object
type ListEnvironmentVariablesV2Request struct {
	ProjectId     string  `json:"project_id"`
	InstanceId    string  `json:"instance_id"`
	GroupId       string  `json:"group_id"`
	EnvId         *string `json:"env_id,omitempty"`
	VariableName  *string `json:"variable_name,omitempty"`
	Offset        *int64  `json:"offset,omitempty"`
	Limit         *int32  `json:"limit,omitempty"`
	PreciseSearch *string `json:"precise_search,omitempty"`
}

func (o ListEnvironmentVariablesV2Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListEnvironmentVariablesV2Request struct{}"
	}

	return strings.Join([]string{"ListEnvironmentVariablesV2Request", string(data)}, " ")
}

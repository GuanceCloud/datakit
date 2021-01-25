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
type ListApisBindedToAppV2Request struct {
	ProjectId  string  `json:"project_id"`
	InstanceId string  `json:"instance_id"`
	AppId      string  `json:"app_id"`
	ApiId      *string `json:"api_id,omitempty"`
	ApiName    *string `json:"api_name,omitempty"`
	GroupId    *string `json:"group_id,omitempty"`
	GroupName  *string `json:"group_name,omitempty"`
	EnvId      *string `json:"env_id,omitempty"`
	Offset     *int64  `json:"offset,omitempty"`
	Limit      *int32  `json:"limit,omitempty"`
}

func (o ListApisBindedToAppV2Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListApisBindedToAppV2Request struct{}"
	}

	return strings.Join([]string{"ListApisBindedToAppV2Request", string(data)}, " ")
}

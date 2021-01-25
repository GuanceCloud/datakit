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

type AppAuthUnBindedApiResp struct {
	// API的认证方式
	AuthType *string `json:"auth_type,omitempty"`
	// 发布的环境名
	RunEnvName *string `json:"run_env_name,omitempty"`
	// API所属分组的名称
	GroupName *string `json:"group_name,omitempty"`
	// 发布记录的编号
	PublishId *string `json:"publish_id,omitempty"`
	// API所属分组的编号
	GroupId *string `json:"group_id,omitempty"`
	// API名称
	Name *string `json:"name,omitempty"`
	// API描述
	Remark *string `json:"remark,omitempty"`
	// 发布的环境id
	RunEnvId *string `json:"run_env_id,omitempty"`
	// API编号
	Id *string `json:"id,omitempty"`
	// API的访问地址
	ReqUri *string `json:"req_uri,omitempty"`
}

func (o AppAuthUnBindedApiResp) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AppAuthUnBindedApiResp struct{}"
	}

	return strings.Join([]string{"AppAuthUnBindedApiResp", string(data)}, " ")
}

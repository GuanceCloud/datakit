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

// 部署资源。
type ReferResources struct {
	// 资源ID。
	Id   *string       `json:"id,omitempty"`
	Type *ResourceType `json:"type,omitempty"`
	// 应用别名，dcs时才提供，支持“distributed_session”、“distributed_cache”、“distributed_session, distributed_cache”，  默认值是“distributed_session, distributed_cache”。
	ReferAlias *string `json:"refer_alias,omitempty"`
	// 引用资源参数。
	Parameters *interface{} `json:"parameters,omitempty"`
}

func (o ReferResources) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ReferResources struct{}"
	}

	return strings.Join([]string{"ReferResources", string(data)}, " ")
}

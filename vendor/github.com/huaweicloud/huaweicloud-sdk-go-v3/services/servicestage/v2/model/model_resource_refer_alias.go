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

// 应用别名，dcs时才提供，支持“distributed_session”、“distributed_cache”、“distributed_session, distributed_cache”，  默认值是“distributed_session, distributed_cache”。
type ResourceReferAlias struct {
}

func (o ResourceReferAlias) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResourceReferAlias struct{}"
	}

	return strings.Join([]string{"ResourceReferAlias", string(data)}, " ")
}

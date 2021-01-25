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

type NamespacesNamespaces struct {
	// 命名空间ID。
	Id string `json:"id"`
	// 命名空间名称。
	Name string `json:"name"`
}

func (o NamespacesNamespaces) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "NamespacesNamespaces struct{}"
	}

	return strings.Join([]string{"NamespacesNamespaces", string(data)}, " ")
}

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

// Response Object
type ListNamespacesResponse struct {
	// 命名空间列表。
	Namespaces     *[]NamespacesNamespaces `json:"namespaces,omitempty"`
	HttpStatusCode int                     `json:"-"`
}

func (o ListNamespacesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListNamespacesResponse struct{}"
	}

	return strings.Join([]string{"ListNamespacesResponse", string(data)}, " ")
}

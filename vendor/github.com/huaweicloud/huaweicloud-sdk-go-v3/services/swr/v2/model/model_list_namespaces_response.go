/*
 * SWR
 *
 * SWR API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListNamespacesResponse struct {
	// 组织列表
	Namespaces     *[]ShowNamespace `json:"namespaces,omitempty"`
	HttpStatusCode int              `json:"-"`
}

func (o ListNamespacesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListNamespacesResponse struct{}"
	}

	return strings.Join([]string{"ListNamespacesResponse", string(data)}, " ")
}

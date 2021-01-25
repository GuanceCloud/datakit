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
type ListInstancesResponse struct {
	// 实例总数。
	Count *int32 `json:"count,omitempty"`
	// 实例列表。
	Instances      *[]InstanceListView `json:"instances,omitempty"`
	HttpStatusCode int                 `json:"-"`
}

func (o ListInstancesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListInstancesResponse struct{}"
	}

	return strings.Join([]string{"ListInstancesResponse", string(data)}, " ")
}

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
type ListApplicationsResponse struct {
	// 应用总数。
	Count *int32 `json:"count,omitempty"`
	// 应用列表。
	Applications   *[]ApplicationView `json:"applications,omitempty"`
	HttpStatusCode int                `json:"-"`
}

func (o ListApplicationsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListApplicationsResponse struct{}"
	}

	return strings.Join([]string{"ListApplicationsResponse", string(data)}, " ")
}

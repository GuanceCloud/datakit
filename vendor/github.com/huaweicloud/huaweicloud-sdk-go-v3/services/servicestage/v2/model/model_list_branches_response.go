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
type ListBranchesResponse struct {
	// 项目分支列表。
	Branches       *[]string `json:"branches,omitempty"`
	HttpStatusCode int       `json:"-"`
}

func (o ListBranchesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListBranchesResponse struct{}"
	}

	return strings.Join([]string{"ListBranchesResponse", string(data)}, " ")
}

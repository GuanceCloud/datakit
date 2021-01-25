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

// Request Object
type ListTreesRequest struct {
	XRepoAuth string `json:"X-Repo-Auth"`
	Namespace string `json:"namespace"`
	Project   string `json:"project"`
	Ref       string `json:"ref"`
}

func (o ListTreesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListTreesRequest struct{}"
	}

	return strings.Join([]string{"ListTreesRequest", string(data)}, " ")
}

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
type ListHooksRequest struct {
	XRepoAuth string `json:"X-Repo-Auth"`
	Namespace string `json:"namespace"`
	Project   string `json:"project"`
}

func (o ListHooksRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListHooksRequest struct{}"
	}

	return strings.Join([]string{"ListHooksRequest", string(data)}, " ")
}

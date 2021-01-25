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
type DeleteHookRequest struct {
	XRepoAuth string `json:"X-Repo-Auth"`
	Namespace string `json:"namespace"`
	Project   string `json:"project"`
	HookId    string `json:"hook_id"`
}

func (o DeleteHookRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteHookRequest struct{}"
	}

	return strings.Join([]string{"DeleteHookRequest", string(data)}, " ")
}

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
type CreateHookRequest struct {
	XRepoAuth string      `json:"X-Repo-Auth"`
	Namespace string      `json:"namespace"`
	Project   string      `json:"project"`
	Body      *HookCreate `json:"body,omitempty"`
}

func (o CreateHookRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateHookRequest struct{}"
	}

	return strings.Join([]string{"CreateHookRequest", string(data)}, " ")
}

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
type ShowContentRequest struct {
	XRepoAuth string `json:"X-Repo-Auth"`
	Namespace string `json:"namespace"`
	Project   string `json:"project"`
	Path      string `json:"path"`
	Ref       string `json:"ref"`
}

func (o ShowContentRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowContentRequest struct{}"
	}

	return strings.Join([]string{"ShowContentRequest", string(data)}, " ")
}

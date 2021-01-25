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
type ShowProjectDetailRequest struct {
	XRepoAuth string `json:"X-Repo-Auth"`
	CloneUrl  string `json:"clone_url"`
}

func (o ShowProjectDetailRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowProjectDetailRequest struct{}"
	}

	return strings.Join([]string{"ShowProjectDetailRequest", string(data)}, " ")
}

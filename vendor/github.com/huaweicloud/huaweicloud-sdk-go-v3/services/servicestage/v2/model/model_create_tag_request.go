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
type CreateTagRequest struct {
	XRepoAuth string     `json:"X-Repo-Auth"`
	Namespace string     `json:"namespace"`
	Project   string     `json:"project"`
	Ref       string     `json:"ref"`
	Body      *TagCreate `json:"body,omitempty"`
}

func (o CreateTagRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateTagRequest struct{}"
	}

	return strings.Join([]string{"CreateTagRequest", string(data)}, " ")
}

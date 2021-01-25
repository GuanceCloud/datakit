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
type CreateProjectRequest struct {
	XRepoAuth string         `json:"X-Repo-Auth"`
	Namespace string         `json:"namespace"`
	Body      *ProjectCreate `json:"body,omitempty"`
}

func (o CreateProjectRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateProjectRequest struct{}"
	}

	return strings.Join([]string{"CreateProjectRequest", string(data)}, " ")
}

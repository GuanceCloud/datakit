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
type CreateFileRequest struct {
	XRepoAuth string      `json:"X-Repo-Auth"`
	Namespace string      `json:"namespace"`
	Project   string      `json:"project"`
	Path      string      `json:"path"`
	Ref       string      `json:"ref"`
	Body      *FileCreate `json:"body,omitempty"`
}

func (o CreateFileRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateFileRequest struct{}"
	}

	return strings.Join([]string{"CreateFileRequest", string(data)}, " ")
}

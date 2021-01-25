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
type UpdateFileRequest struct {
	XRepoAuth string      `json:"X-Repo-Auth"`
	Namespace string      `json:"namespace"`
	Project   string      `json:"project"`
	Path      string      `json:"path"`
	Ref       string      `json:"ref"`
	Body      *FileUpdate `json:"body,omitempty"`
}

func (o UpdateFileRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateFileRequest struct{}"
	}

	return strings.Join([]string{"UpdateFileRequest", string(data)}, " ")
}

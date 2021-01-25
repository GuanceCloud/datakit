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
type DeleteFileRequest struct {
	XRepoAuth string `json:"X-Repo-Auth"`
	Namespace string `json:"namespace"`
	Project   string `json:"project"`
	Path      string `json:"path"`
	Ref       string `json:"ref"`
	Message   string `json:"message"`
	Sha       string `json:"sha"`
}

func (o DeleteFileRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteFileRequest struct{}"
	}

	return strings.Join([]string{"DeleteFileRequest", string(data)}, " ")
}

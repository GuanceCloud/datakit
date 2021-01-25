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
type ListCommitsRequest struct {
	XRepoAuth string  `json:"X-Repo-Auth"`
	Namespace string  `json:"namespace"`
	Project   string  `json:"project"`
	Ref       *string `json:"ref,omitempty"`
}

func (o ListCommitsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListCommitsRequest struct{}"
	}

	return strings.Join([]string{"ListCommitsRequest", string(data)}, " ")
}

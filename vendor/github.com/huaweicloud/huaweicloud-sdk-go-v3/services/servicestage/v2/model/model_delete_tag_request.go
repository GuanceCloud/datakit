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
type DeleteTagRequest struct {
	XRepoAuth string `json:"X-Repo-Auth"`
	Namespace string `json:"namespace"`
	Project   string `json:"project"`
	TagName   string `json:"tag_name"`
}

func (o DeleteTagRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteTagRequest struct{}"
	}

	return strings.Join([]string{"DeleteTagRequest", string(data)}, " ")
}

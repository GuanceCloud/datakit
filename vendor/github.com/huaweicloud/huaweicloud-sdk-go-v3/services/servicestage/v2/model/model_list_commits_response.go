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

// Response Object
type ListCommitsResponse struct {
	// 提交记录列表。
	Commits        *[]CommitsCommits `json:"commits,omitempty"`
	HttpStatusCode int               `json:"-"`
}

func (o ListCommitsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListCommitsResponse struct{}"
	}

	return strings.Join([]string{"ListCommitsResponse", string(data)}, " ")
}

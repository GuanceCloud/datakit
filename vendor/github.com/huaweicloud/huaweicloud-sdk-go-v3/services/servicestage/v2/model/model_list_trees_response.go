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
type ListTreesResponse struct {
	// 仓库文件列表。
	Paths          *[]string `json:"paths,omitempty"`
	HttpStatusCode int       `json:"-"`
}

func (o ListTreesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListTreesResponse struct{}"
	}

	return strings.Join([]string{"ListTreesResponse", string(data)}, " ")
}

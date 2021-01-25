/*
 * SWR
 *
 * SWR API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListRepositoryTagsResponse struct {
	// 镜像tag列表
	Body           *[]ShowReposTagResp `json:"body,omitempty"`
	HttpStatusCode int                 `json:"-"`
}

func (o ListRepositoryTagsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListRepositoryTagsResponse struct{}"
	}

	return strings.Join([]string{"ListRepositoryTagsResponse", string(data)}, " ")
}

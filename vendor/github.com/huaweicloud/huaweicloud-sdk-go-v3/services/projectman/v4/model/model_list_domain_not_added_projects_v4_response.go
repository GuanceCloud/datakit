/*
 * ProjectMan
 *
 * devcloud projectman api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListDomainNotAddedProjectsV4Response struct {
	// 项目信息列表
	Projects *[]ListDomainNotAddedProjectsV4ResponseBodyProjects `json:"projects,omitempty"`
	// 总数
	Total          *int32 `json:"total,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListDomainNotAddedProjectsV4Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListDomainNotAddedProjectsV4Response struct{}"
	}

	return strings.Join([]string{"ListDomainNotAddedProjectsV4Response", string(data)}, " ")
}

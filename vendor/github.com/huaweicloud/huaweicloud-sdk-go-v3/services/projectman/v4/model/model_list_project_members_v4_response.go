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
type ListProjectMembersV4Response struct {
	// 项目成员列表
	Members *[]MemberListV4Members `json:"members,omitempty"`
	// 总数
	Total          *int32 `json:"total,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListProjectMembersV4Response) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListProjectMembersV4Response struct{}"
	}

	return strings.Join([]string{"ListProjectMembersV4Response", string(data)}, " ")
}

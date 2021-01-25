/*
 * Classroom
 *
 * devcloud classedge api
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListClassroomMembersResponse struct {
	// 课堂成员列表
	Members *[]ClassroomMember `json:"members,omitempty"`
	// 课堂成员总数
	Total          *int32 `json:"total,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListClassroomMembersResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListClassroomMembersResponse struct{}"
	}

	return strings.Join([]string{"ListClassroomMembersResponse", string(data)}, " ")
}

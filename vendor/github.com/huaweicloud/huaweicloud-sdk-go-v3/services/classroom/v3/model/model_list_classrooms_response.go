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
type ListClassroomsResponse struct {
	// 课堂列表
	Classrooms *[]ClassroomCard `json:"classrooms,omitempty"`
	// 课堂总数
	Total          *int32 `json:"total,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListClassroomsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListClassroomsResponse struct{}"
	}

	return strings.Join([]string{"ListClassroomsResponse", string(data)}, " ")
}

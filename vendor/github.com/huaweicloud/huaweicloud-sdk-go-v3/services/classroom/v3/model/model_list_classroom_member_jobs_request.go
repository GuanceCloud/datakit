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

// Request Object
type ListClassroomMemberJobsRequest struct {
	ClassroomId string `json:"classroom_id"`
	MemberId    string `json:"member_id"`
	Offset      *int32 `json:"offset,omitempty"`
	Limit       *int32 `json:"limit,omitempty"`
}

func (o ListClassroomMemberJobsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListClassroomMemberJobsRequest struct{}"
	}

	return strings.Join([]string{"ListClassroomMemberJobsRequest", string(data)}, " ")
}

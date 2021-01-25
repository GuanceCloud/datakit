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
type ListMemberJobRecordsRequest struct {
	JobId      string `json:"job_id"`
	ExerciseId string `json:"exercise_id"`
	MemberId   string `json:"member_id"`
	Offset     *int32 `json:"offset,omitempty"`
	Limit      *int32 `json:"limit,omitempty"`
}

func (o ListMemberJobRecordsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListMemberJobRecordsRequest struct{}"
	}

	return strings.Join([]string{"ListMemberJobRecordsRequest", string(data)}, " ")
}

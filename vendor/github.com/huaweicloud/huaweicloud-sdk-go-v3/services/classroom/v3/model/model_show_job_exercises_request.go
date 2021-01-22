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
type ShowJobExercisesRequest struct {
	JobId      string `json:"job_id"`
	SourceFrom string `json:"source_from"`
	SourceId   string `json:"source_id"`
	Offset     *int32 `json:"offset,omitempty"`
	Limit      *int32 `json:"limit,omitempty"`
}

func (o ShowJobExercisesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowJobExercisesRequest struct{}"
	}

	return strings.Join([]string{"ShowJobExercisesRequest", string(data)}, " ")
}

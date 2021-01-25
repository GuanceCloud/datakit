/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ListRestoreRecordsRequest struct {
	InstanceId string  `json:"instance_id"`
	BeginTime  *string `json:"begin_time,omitempty"`
	EndTime    *string `json:"end_time,omitempty"`
	Limit      *int32  `json:"limit,omitempty"`
	Offset     *int32  `json:"offset,omitempty"`
}

func (o ListRestoreRecordsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListRestoreRecordsRequest struct{}"
	}

	return strings.Join([]string{"ListRestoreRecordsRequest", string(data)}, " ")
}

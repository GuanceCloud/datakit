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
type ListMemberJobRecordsResponse struct {
	// 习题提交列表信息
	Records *[]JobRecords `json:"records,omitempty"`
	// 习题提交总次数
	Total          *int32 `json:"total,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListMemberJobRecordsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListMemberJobRecordsResponse struct{}"
	}

	return strings.Join([]string{"ListMemberJobRecordsResponse", string(data)}, " ")
}

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

// Response Object
type ListRestoreRecordsResponse struct {
	// 实例恢复记录的详情数组。
	RestoreRecordResponse *[]InstanceRestoreInfo `json:"restore_record_response,omitempty"`
	// 返回记录数。
	TotalNum       *int32 `json:"total_num,omitempty"`
	HttpStatusCode int    `json:"-"`
}

func (o ListRestoreRecordsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListRestoreRecordsResponse struct{}"
	}

	return strings.Join([]string{"ListRestoreRecordsResponse", string(data)}, " ")
}

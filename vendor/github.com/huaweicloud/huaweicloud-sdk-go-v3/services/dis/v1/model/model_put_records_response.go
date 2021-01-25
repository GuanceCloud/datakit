/*
 * DIS
 *
 * DIS v1 API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type PutRecordsResponse struct {
	// 上传失败的数据数量。
	FailedRecordCount *int32                   `json:"failed_record_count,omitempty"`
	Records           *[]PutRecordsResultEntry `json:"records,omitempty"`
	HttpStatusCode    int                      `json:"-"`
}

func (o PutRecordsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PutRecordsResponse struct{}"
	}

	return strings.Join([]string{"PutRecordsResponse", string(data)}, " ")
}

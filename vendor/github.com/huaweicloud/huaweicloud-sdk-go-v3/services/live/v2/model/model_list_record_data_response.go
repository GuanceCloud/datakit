/*
 * Live
 *
 * 数据分析服务接口
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListRecordDataResponse struct {
	// 采样数据列表。
	RecordDataList *[]RecordData `json:"record_data_list,omitempty"`
	XRequestId     *string       `json:"X-request-id,omitempty"`
	HttpStatusCode int           `json:"-"`
}

func (o ListRecordDataResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListRecordDataResponse struct{}"
	}

	return strings.Join([]string{"ListRecordDataResponse", string(data)}, " ")
}

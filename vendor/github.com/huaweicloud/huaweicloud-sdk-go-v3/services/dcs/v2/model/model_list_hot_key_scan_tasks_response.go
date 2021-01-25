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
type ListHotKeyScanTasksResponse struct {
	// 实例ID
	InstanceId *string `json:"instance_id,omitempty"`
	// 总数
	Count *int32 `json:"count,omitempty"`
	// 热key分析记录列表
	Records        *[]RecordsResponse `json:"records,omitempty"`
	HttpStatusCode int                `json:"-"`
}

func (o ListHotKeyScanTasksResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListHotKeyScanTasksResponse struct{}"
	}

	return strings.Join([]string{"ListHotKeyScanTasksResponse", string(data)}, " ")
}

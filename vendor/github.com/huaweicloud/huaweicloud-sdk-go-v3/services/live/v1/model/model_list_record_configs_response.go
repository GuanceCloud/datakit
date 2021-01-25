/*
 * Live
 *
 * 直播服务源站所有接口
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ListRecordConfigsResponse struct {
	// 查询结果的总元素数量
	Total *int32 `json:"total,omitempty"`
	// 录制配置数组
	RecordConfig   *[]RecordConfigInfo `json:"record_config,omitempty"`
	HttpStatusCode int                 `json:"-"`
}

func (o ListRecordConfigsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListRecordConfigsResponse struct{}"
	}

	return strings.Join([]string{"ListRecordConfigsResponse", string(data)}, " ")
}

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
type ListSnapshotDataResponse struct {
	// 采样数据列表。
	SnapshotList *[]SnapshotData `json:"snapshot_list,omitempty"`
	// 指定时间区间内截图数量总和
	Total          *int64  `json:"total,omitempty"`
	XRequestId     *string `json:"X-request-id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ListSnapshotDataResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListSnapshotDataResponse struct{}"
	}

	return strings.Join([]string{"ListSnapshotDataResponse", string(data)}, " ")
}

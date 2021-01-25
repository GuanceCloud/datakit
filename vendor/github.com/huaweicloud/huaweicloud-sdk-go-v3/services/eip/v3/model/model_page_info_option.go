/*
 * EIP
 *
 * 云服务接口
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 分页页码信息
type PageInfoOption struct {
	// 翻页时，作为前一页的marker取值
	PreviousMarker *string `json:"previous_marker,omitempty"`
	// 翻页时，作为后一页的marker取值
	NextMarker *string `json:"next_marker,omitempty"`
	// 当前页的数据总数
	CurrentCount *int32 `json:"current_count,omitempty"`
}

func (o PageInfoOption) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PageInfoOption struct{}"
	}

	return strings.Join([]string{"PageInfoOption", string(data)}, " ")
}

/*
 * VPC
 *
 * VPC Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

//
type PageInfo struct {
	// 当前页第一条记录
	PreviousMarker string `json:"previous_marker"`
	// 当前页总数
	CurrentCount int32 `json:"current_count"`
	// 当前页最后一条记录，最后一页时无next_marker字段
	NextMarker *string `json:"next_marker,omitempty"`
}

func (o PageInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PageInfo struct{}"
	}

	return strings.Join([]string{"PageInfo", string(data)}, " ")
}

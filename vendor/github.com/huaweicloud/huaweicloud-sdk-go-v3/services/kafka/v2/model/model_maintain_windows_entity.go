/*
 * Kafka
 *
 * Kafka Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type MaintainWindowsEntity struct {
	// 是否为默认时间段。
	Default *bool `json:"default,omitempty"`
	// 维护时间窗结束时间。
	End *string `json:"end,omitempty"`
	// 维护时间窗开始时间。
	Begin *string `json:"begin,omitempty"`
	// 序号。
	Seq *int32 `json:"seq,omitempty"`
}

func (o MaintainWindowsEntity) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "MaintainWindowsEntity struct{}"
	}

	return strings.Join([]string{"MaintainWindowsEntity", string(data)}, " ")
}

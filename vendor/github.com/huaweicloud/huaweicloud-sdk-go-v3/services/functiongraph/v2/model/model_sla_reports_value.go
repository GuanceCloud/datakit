/*
 * FunctionGraph
 *
 * API v2
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type SlaReportsValue struct {
	// 时间戳
	Timestamp *int32 `json:"timestamp,omitempty"`
	// 值
	Value *int32 `json:"value,omitempty"`
}

func (o SlaReportsValue) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SlaReportsValue struct{}"
	}

	return strings.Join([]string{"SlaReportsValue", string(data)}, " ")
}

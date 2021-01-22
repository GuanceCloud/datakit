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

type MonthUsed struct {
	// 日期
	Date *string `json:"date,omitempty"`
	// 使用量
	Value *int32 `json:"value,omitempty"`
}

func (o MonthUsed) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "MonthUsed struct{}"
	}

	return strings.Join([]string{"MonthUsed", string(data)}, " ")
}

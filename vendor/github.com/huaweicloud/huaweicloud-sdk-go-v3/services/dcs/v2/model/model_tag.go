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

type Tag struct {
	// 标签键，最大长度36个unicode字符。
	Key *string `json:"key,omitempty"`
	// 标签值
	Values *[]string `json:"values,omitempty"`
}

func (o Tag) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Tag struct{}"
	}

	return strings.Join([]string{"Tag", string(data)}, " ")
}

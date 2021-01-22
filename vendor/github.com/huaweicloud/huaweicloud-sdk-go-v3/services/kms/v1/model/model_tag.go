/*
 * kms
 *
 * KMS v1.0 API, open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type Tag struct {
	// 键。 最大长度36个unicode字符。 key不能为空。不能包含非打印字符“ASCII(0-31)”、“*”、“<”、“>”、“\\”、“=”。
	Key *string `json:"key,omitempty"`
	// 标签值集合
	Values *[]string `json:"values,omitempty"`
}

func (o Tag) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Tag struct{}"
	}

	return strings.Join([]string{"Tag", string(data)}, " ")
}

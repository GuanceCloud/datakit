/*
 * DIS
 *
 * DIS v1 API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type Tag struct {
	// 键。  - 不能为空。  - 对于同一资源键值唯一。  - 字符集：A-Z，a-z ， 0-9，‘-’，‘_’，UNICODE字符（\\u4E00-\\u9FFF）。
	Key *string `json:"key,omitempty"`
	// 值。  - 长度不超过43个字符。  - 字符集：A-Z，a-z ， 0-9，‘.’，‘-’，‘_’，UNICODE字符（\\u4E00-\\u9FFF）。  - 只能包含数字、字母、中划线“-”、下划线“_”。
	Value *string `json:"value,omitempty"`
}

func (o Tag) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Tag struct{}"
	}

	return strings.Join([]string{"Tag", string(data)}, " ")
}

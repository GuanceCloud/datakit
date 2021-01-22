/*
 * RDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type InstanceRequestTags struct {
	// 标签键。创建时，最大长度36个unicode字符，key不能为空，不能为空字符串，不能重复。字符集：A-Z，a-z ， 0-9，‘-’，‘_’，UNICODE字符（\\u4E00-\\u9FFF）；删除时，最大长度127个unicode字符，key不能为空，不能为空字符串。
	Key string `json:"key"`
	// 标签值。创建时，每个值最大长度43个unicode字符，可以为空字符串。 字符集：A-Z，a-z ， 0-9，‘.’，‘-’，‘_’，UNICODE字符（\\u4E00-\\u9FFF）；删除时，每个值最大长度255个unicode字符，如果value有值按照key/value删除，如果value没值则按照key删除。
	Value string `json:"value"`
}

func (o InstanceRequestTags) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "InstanceRequestTags struct{}"
	}

	return strings.Join([]string{"InstanceRequestTags", string(data)}, " ")
}

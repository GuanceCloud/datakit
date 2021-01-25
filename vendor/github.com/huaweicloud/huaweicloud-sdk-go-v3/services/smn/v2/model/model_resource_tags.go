/*
 * SMN
 *
 * SMN Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 资源标签列表结构体。
type ResourceTags struct {
	// 键。  - 最大长度127个unicode字符。  - key不能为空。
	Key string `json:"key"`
	// 值列表。  - 最多10个value。  - value不允许重复。  - 每个值最大长度255个unicode字符。  - 如果values为空则表示any_value。  - value之间为或的关系。
	Values []string `json:"values"`
}

func (o ResourceTags) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResourceTags struct{}"
	}

	return strings.Join([]string{"ResourceTags", string(data)}, " ")
}

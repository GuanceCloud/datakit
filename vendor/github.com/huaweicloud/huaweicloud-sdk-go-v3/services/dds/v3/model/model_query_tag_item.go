/*
 * DDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type QueryTagItem struct {
	// 标签键。最大长度36个unicode字符，key不能为空或者空字符串，不能为空格，校验和使用之前先排除前后半角空格。
	Key string `json:"key"`
	// 标签值列表。最大长度43个unicode字符，不能为空格，校验和使用之前先排除前后半角空格。如果values为空列表，则表示查询任意value。value之间为或的关系。
	Values []string `json:"values"`
}

func (o QueryTagItem) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "QueryTagItem struct{}"
	}

	return strings.Join([]string{"QueryTagItem", string(data)}, " ")
}

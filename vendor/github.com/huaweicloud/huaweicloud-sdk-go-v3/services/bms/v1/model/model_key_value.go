/*
 * BMS
 *
 * BMS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// metadata数据结构说明
type KeyValue struct {
	// 键。最大长度255个Unicode字符，不能为空。可以为大写字母（A-Z）、小写字母（a-z）、数字（0-9）、中划线（-）、下划线（_）、冒号（:）和小数点（.）。
	Key string `json:"key"`
}

func (o KeyValue) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "KeyValue struct{}"
	}

	return strings.Join([]string{"KeyValue", string(data)}, " ")
}

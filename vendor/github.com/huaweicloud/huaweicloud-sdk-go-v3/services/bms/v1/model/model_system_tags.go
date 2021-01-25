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

// 裸金属服务器的标签。详情请参见表 server_tags字段数据结构说明。 说明：创建裸金属服务器时，一台裸金属服务器最多可以添加10个标签。其中，__type_baremetal为系统内部标签，因此实际能添加的标签为9个。
type SystemTags struct {
	// 键。最大长度36个unicode字符。key不能为空。不能包含非打印字符ASCII（0-31），以及特殊字符同一资源的key值不能重复。
	Key *string `json:"key,omitempty"`
	// 值。每个值最大长度43个unicode字符，可以为空字符串。不能包含非打印字符ASCII（0-31），以及特殊字符
	Value *string `json:"value,omitempty"`
}

func (o SystemTags) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SystemTags struct{}"
	}

	return strings.Join([]string{"SystemTags", string(data)}, " ")
}

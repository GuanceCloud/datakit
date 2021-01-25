/*
 * EIP
 *
 * 云服务接口
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 标签信息
type TagsInfo struct {
	// 功能说明：键。同一资源的key值不能重复。
	Key *string `json:"key,omitempty"`
	// 功能说明：值列表。
	Value *string `json:"value,omitempty"`
}

func (o TagsInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "TagsInfo struct{}"
	}

	return strings.Join([]string{"TagsInfo", string(data)}, " ")
}

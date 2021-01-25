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

// 标签
type ResourceTagResp struct {
	// 键。同一资源的key值不能重复。
	Key *string `json:"key,omitempty"`
	// 值列表。
	Value *string `json:"value,omitempty"`
}

func (o ResourceTagResp) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResourceTagResp struct{}"
	}

	return strings.Join([]string{"ResourceTagResp", string(data)}, " ")
}

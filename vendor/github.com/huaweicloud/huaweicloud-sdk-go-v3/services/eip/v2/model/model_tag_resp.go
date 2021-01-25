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
type TagResp struct {
	// 键，key不能为空。长度不超过36个字符。由英文字母、数字、下划线、中划线、中文字符组成。
	Key *string `json:"key,omitempty"`
	// 值列表。
	Values *[]string `json:"values,omitempty"`
}

func (o TagResp) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "TagResp struct{}"
	}

	return strings.Join([]string{"TagResp", string(data)}, " ")
}

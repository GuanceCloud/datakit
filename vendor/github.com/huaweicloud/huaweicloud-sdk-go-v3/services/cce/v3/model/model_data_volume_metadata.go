/*
 * CCE
 *
 * CCE开放API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 数据盘加密信息，仅在创建节点数据盘需加密时须填写。
type DataVolumeMetadata struct {
	// 用户主密钥ID，是metadata中的表示加密功能的字段，与__system__encrypted配合使用。
	SystemCmkid *string `json:"__system__cmkid,omitempty"`
	// 表示云硬盘加密功能的字段，'0'代表不加密，'1'代表加密。  该字段不存在时，云硬盘默认为不加密。
	SystemEncrypted *string `json:"__system__encrypted,omitempty"`
}

func (o DataVolumeMetadata) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DataVolumeMetadata struct{}"
	}

	return strings.Join([]string{"DataVolumeMetadata", string(data)}, " ")
}

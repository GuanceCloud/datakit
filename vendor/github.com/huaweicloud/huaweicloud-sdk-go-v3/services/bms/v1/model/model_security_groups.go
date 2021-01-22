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

// security_groups字段数据结构说明
type SecurityGroups struct {
	// 创建裸金属服务器时未指定安全组，该值为default。创建裸金属服务器时，需要指定已有安全组的ID（UUID格式）。
	Name *string `json:"name,omitempty"`
}

func (o SecurityGroups) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SecurityGroups struct{}"
	}

	return strings.Join([]string{"SecurityGroups", string(data)}, " ")
}

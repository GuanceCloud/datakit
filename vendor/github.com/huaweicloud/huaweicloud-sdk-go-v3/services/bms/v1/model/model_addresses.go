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

// addresses字段数据结构说明
type Addresses struct {
	// 裸金属服务器所属网络信息。key表示裸金属服务器使用的虚拟私有云的ID。value为网络详细信息
	VpcId []Address `json:"vpc_id"`
}

func (o Addresses) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Addresses struct{}"
	}

	return strings.Join([]string{"Addresses", string(data)}, " ")
}

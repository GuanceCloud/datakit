/*
 * RDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 单机转主备时必填。
type Single2Ha struct {
	// 实例节点可用区码（AZ）。
	AzCodeNewNode string `json:"az_code_new_node"`
	// 仅在支持SQL Server数据库实例进行单机转主备时必选，有效。
	Password *string `json:"password,omitempty"`
	// 创建新节点所在专属存储池ID，仅专属云创建实例时有效。
	DsspoolId *string `json:"dsspool_id,omitempty"`
}

func (o Single2Ha) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "Single2Ha struct{}"
	}

	return strings.Join([]string{"Single2Ha", string(data)}, " ")
}

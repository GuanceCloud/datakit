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

// 跨区域备份实例信息。
type OffsiteBackupInstance struct {
	// 偏移量。
	Offset string `json:"offset"`
	// 查询记录数。
	Limit string `json:"limit"`
}

func (o OffsiteBackupInstance) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "OffsiteBackupInstance struct{}"
	}

	return strings.Join([]string{"OffsiteBackupInstance", string(data)}, " ")
}

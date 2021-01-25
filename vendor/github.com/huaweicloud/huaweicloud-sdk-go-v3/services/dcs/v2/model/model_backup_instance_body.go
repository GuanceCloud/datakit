/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 备份实例请求体
type BackupInstanceBody struct {
	// 备份缓存实例的备注信息。
	Remark *string `json:"remark,omitempty"`
}

func (o BackupInstanceBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BackupInstanceBody struct{}"
	}

	return strings.Join([]string{"BackupInstanceBody", string(data)}, " ")
}

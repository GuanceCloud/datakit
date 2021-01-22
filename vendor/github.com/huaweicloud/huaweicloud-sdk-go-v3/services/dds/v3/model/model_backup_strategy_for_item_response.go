/*
 * DDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 自动备份策略。
type BackupStrategyForItemResponse struct {
	// 备份时间段。自动备份将在该时间段内触发。当前时间指UTC时间。
	StartTime string `json:"start_time"`
	// 已生成备份文件可以保存的天数。取值范围：0~732。
	KeepDays int32 `json:"keep_days"`
}

func (o BackupStrategyForItemResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BackupStrategyForItemResponse struct{}"
	}

	return strings.Join([]string{"BackupStrategyForItemResponse", string(data)}, " ")
}

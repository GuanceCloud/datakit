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

type BackupPlan struct {
	// 备份的时区。取值为-1200 ~+1200之间的时区。若为空则默认使用DCS-Server节点的当前时区。
	TimezoneOffset *string `json:"timezone_offset,omitempty"`
	// 每周的周几开始备份，取值1-7，1代表周一，7代表周日。
	BackupAt []int32 `json:"backup_at"`
	// 备份周期类型，目前支持“weekly”。
	PeriodType string `json:"period_type"`
	// 备份执行时间，“00:00-01:00”代表0点开始执行备份。
	BeginAt string `json:"begin_at"`
}

func (o BackupPlan) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BackupPlan struct{}"
	}

	return strings.Join([]string{"BackupPlan", string(data)}, " ")
}

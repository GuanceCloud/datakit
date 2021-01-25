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

type BackupPolicy struct {
	// 备份类型。 - auto：自动备份 - manual：手动备份
	BackupType *string `json:"backup_type,omitempty"`
	// 当backup_type设置为manual时，该参数为必填。 保留天数，单位：天，取值范围：1-7。
	SaveDays             *int32      `json:"save_days,omitempty"`
	PeriodicalBackupPlan *BackupPlan `json:"periodical_backup_plan"`
}

func (o BackupPolicy) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BackupPolicy struct{}"
	}

	return strings.Join([]string{"BackupPolicy", string(data)}, " ")
}

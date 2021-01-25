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

// 跨区域备份策略信息。
type OffSiteBackupPolicy struct {
	// 指定已生成的备份文件可以保存的天数。  取值范围：0～732。取0值，表示关闭自动备份策略。如果需要延长保留时间请联系客服人员申请，自动备份最长可以保留2562天。  注意： 关闭备份策略后，备份任务将立即停止，所有增量备份任务将立即删除，使用增量备份的相关操作可能失败，相关操作不限于下载、复制、恢复、重建等，请谨慎操作。
	KeepDays int32 `json:"keep_days"`
	// 备份类型，取值：  - “auto”: 自动全量备份 - “incremental”: 自动增量备份 - “all”: 同时设置自动全量和自动增量备份
	BackupType *interface{} `json:"backup_type,omitempty"`
	// 目标区域ID。
	DestinationRegion *string `json:"destination_region,omitempty"`
	// 项目ID。
	DestinationProjectId *string `json:"destination_project_id,omitempty"`
}

func (o OffSiteBackupPolicy) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "OffSiteBackupPolicy struct{}"
	}

	return strings.Join([]string{"OffSiteBackupPolicy", string(data)}, " ")
}

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

// 实例数量统计信息。
type StatusStatistic struct {
	// 支付中的实例数。
	PayingCount *int32 `json:"paying_count,omitempty"`
	// 冻结中的实例数。
	FreezingCount *int32 `json:"freezing_count,omitempty"`
	// 迁移中的实例数。
	MigratingCount *int32 `json:"migrating_count,omitempty"`
	// 清空中的实例数。
	FlushingCount *int32 `json:"flushing_count,omitempty"`
	// 升级中的实例数。
	UpgradingCount *int32 `json:"upgrading_count,omitempty"`
	// 恢复中的实例数。
	RestoringCount *int32 `json:"restoring_count,omitempty"`
	// 扩容中的实例数。
	ExtendingCount *int32 `json:"extending_count,omitempty"`
	// 正在创建的实例数。
	CreatingCount *int32 `json:"creating_count,omitempty"`
	// 正在运行的实例数。
	RunningCount *int32 `json:"running_count,omitempty"`
	// 异常的实例数。
	ErrorCount *int32 `json:"error_count,omitempty"`
	// 已冻结的实例数。
	FrozenCount *int32 `json:"frozen_count,omitempty"`
	// 创建失败的实例数。
	CreatefailedCount *int32 `json:"createfailed_count,omitempty"`
	// 正在重启的实例数。
	RestartingCount *int32 `json:"restarting_count,omitempty"`
}

func (o StatusStatistic) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "StatusStatistic struct{}"
	}

	return strings.Join([]string{"StatusStatistic", string(data)}, " ")
}

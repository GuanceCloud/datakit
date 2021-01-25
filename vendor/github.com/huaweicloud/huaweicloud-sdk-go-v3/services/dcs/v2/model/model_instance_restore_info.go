/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

type InstanceRestoreInfo struct {
	// 备份记录ID。
	BackupId *string `json:"backup_id,omitempty"`
	// 恢复记录ID。
	RestoreId *string `json:"restore_id,omitempty"`
	// 备份记录名称。
	BackupName *string `json:"backup_name,omitempty"`
	// 恢复完成时间。
	UpdatedAt *string `json:"updated_at,omitempty"`
	// 恢复备注信息。
	RestoreRemark *string `json:"restore_remark,omitempty"`
	// 恢复任务创建时间。
	CreatedAt *string `json:"created_at,omitempty"`
	// 恢复进度。
	Progress *string `json:"progress,omitempty"`
	// 恢复失败后错误码 * `dcs.08.0001` - 启动备份恢复工具失败。 * `dcs.08.0002` - 执行超时。 * `dcs.08.0003` - 删除桶失败。 * `dcs.08.0004` - 获取ak/sk 失败。 * `dcs.08.0005` - 创建桶失败。 * `dcs.08.0006` - 查询备份数据大小失败。 * `dcs.08.0007` - 恢复时同步数据失败。 * `dcs.08.0008` - 自动备份任务未运行，实例正在运行其他任务。
	ErrorCode *string `json:"error_code,omitempty"`
	// 恢复记录名称。
	RestoreName *string `json:"restore_name,omitempty"`
	// 备份备注信息。
	BackupRemark *string `json:"backup_remark,omitempty"`
	// 恢复状态。 - waiting：等待中 - restoring：恢复中 - succeed：恢复成功 - failed：恢复失败
	Status *InstanceRestoreInfoStatus `json:"status,omitempty"`
}

func (o InstanceRestoreInfo) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "InstanceRestoreInfo struct{}"
	}

	return strings.Join([]string{"InstanceRestoreInfo", string(data)}, " ")
}

type InstanceRestoreInfoStatus struct {
	value string
}

type InstanceRestoreInfoStatusEnum struct {
	WAITING   InstanceRestoreInfoStatus
	RESTORING InstanceRestoreInfoStatus
	SUCCEED   InstanceRestoreInfoStatus
	FAILED    InstanceRestoreInfoStatus
}

func GetInstanceRestoreInfoStatusEnum() InstanceRestoreInfoStatusEnum {
	return InstanceRestoreInfoStatusEnum{
		WAITING: InstanceRestoreInfoStatus{
			value: "waiting",
		},
		RESTORING: InstanceRestoreInfoStatus{
			value: "restoring",
		},
		SUCCEED: InstanceRestoreInfoStatus{
			value: "succeed",
		},
		FAILED: InstanceRestoreInfoStatus{
			value: "failed",
		},
	}
}

func (c InstanceRestoreInfoStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *InstanceRestoreInfoStatus) UnmarshalJSON(b []byte) error {
	myConverter := converter.StringConverterFactory("string")
	if myConverter != nil {
		val, err := myConverter.CovertStringToInterface(strings.Trim(string(b[:]), "\""))
		if err == nil {
			c.value = val.(string)
			return nil
		}
		return err
	} else {
		return errors.New("convert enum data to string error")
	}
}

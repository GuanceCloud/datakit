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

// 实例备份信息。
type BackupRecordResponse struct {
	// 备份记录ID。
	BackupId *string `json:"backup_id,omitempty"`
	// 备份执行时间段。
	Period *string `json:"period,omitempty"`
	// 备份记录名称。
	BackupName *string `json:"backup_name,omitempty"`
	// 实例ID。
	InstanceId *string `json:"instance_id,omitempty"`
	// 备份文件大小（Byte）。
	Size *int64 `json:"size,omitempty"`
	// 备份类型。 - manual：表示备份类型为手动备份 - auto：表示备份类型为自动备份
	BackupType *BackupRecordResponseBackupType `json:"backup_type,omitempty"`
	// 备份任务创建时间。
	CreatedAt *string `json:"created_at,omitempty"`
	// 备份完成时间。
	UpdatedAt *string `json:"updated_at,omitempty"`
	// 备份进度。
	Progress *string `json:"progress,omitempty"`
	// 备份失败后错误码 * `dcs.08.0001` - 启动备份恢复工具失败。 * `dcs.08.0002` - 执行超时。 * `dcs.08.0003` - 删除桶失败。 * `dcs.08.0004` - 获取ak/sk 失败。 * `dcs.08.0005` - 创建桶失败。 * `dcs.08.0006` - 查询备份数据大小失败。 * `dcs.08.0007` - 恢复时同步数据失败。 * `dcs.08.0008` - 自动备份任务未运行，实例正在运行其他任务。
	ErrorCode *string `json:"error_code,omitempty"`
	// 备份缓存实例的备注信息。
	Remark *string `json:"remark,omitempty"`
	// 备份状态。 - waiting：等待中。 - backuping：备份中。 - succeed：备份成功。 - failed：备份失败。 - expired：备份文件过期。 - deleted：已手动删除备份文件。
	Status *BackupRecordResponseStatus `json:"status,omitempty"`
	// 是否可以进行恢复操作，取值为TRUE或FALSE。
	IsSupportRestore *string `json:"is_support_restore,omitempty"`
}

func (o BackupRecordResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BackupRecordResponse struct{}"
	}

	return strings.Join([]string{"BackupRecordResponse", string(data)}, " ")
}

type BackupRecordResponseBackupType struct {
	value string
}

type BackupRecordResponseBackupTypeEnum struct {
	MANUAL BackupRecordResponseBackupType
	AUTO   BackupRecordResponseBackupType
}

func GetBackupRecordResponseBackupTypeEnum() BackupRecordResponseBackupTypeEnum {
	return BackupRecordResponseBackupTypeEnum{
		MANUAL: BackupRecordResponseBackupType{
			value: "manual",
		},
		AUTO: BackupRecordResponseBackupType{
			value: "auto",
		},
	}
}

func (c BackupRecordResponseBackupType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BackupRecordResponseBackupType) UnmarshalJSON(b []byte) error {
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

type BackupRecordResponseStatus struct {
	value string
}

type BackupRecordResponseStatusEnum struct {
	WAITING   BackupRecordResponseStatus
	BACKUPING BackupRecordResponseStatus
	SUCCEED   BackupRecordResponseStatus
	FAILED    BackupRecordResponseStatus
	EXPIRED   BackupRecordResponseStatus
	DELETED   BackupRecordResponseStatus
}

func GetBackupRecordResponseStatusEnum() BackupRecordResponseStatusEnum {
	return BackupRecordResponseStatusEnum{
		WAITING: BackupRecordResponseStatus{
			value: "waiting",
		},
		BACKUPING: BackupRecordResponseStatus{
			value: "backuping",
		},
		SUCCEED: BackupRecordResponseStatus{
			value: "succeed",
		},
		FAILED: BackupRecordResponseStatus{
			value: "failed",
		},
		EXPIRED: BackupRecordResponseStatus{
			value: "expired",
		},
		DELETED: BackupRecordResponseStatus{
			value: "deleted",
		},
	}
}

func (c BackupRecordResponseStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BackupRecordResponseStatus) UnmarshalJSON(b []byte) error {
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

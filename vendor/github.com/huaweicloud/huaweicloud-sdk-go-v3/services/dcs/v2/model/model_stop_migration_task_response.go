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

// Response Object
type StopMigrationTaskResponse struct {
	// 迁移任务ID。
	TaskId *string `json:"task_id,omitempty"`
	// 迁移任务名称。
	TaskName *string `json:"task_name,omitempty"`
	// 迁移任务描述。
	Description *string `json:"description,omitempty"`
	// 迁移任务状态，这个字段的值包括：SUCCESS, FAILED, MIGRATING，TERMINATED。
	Status *StopMigrationTaskResponseStatus `json:"status,omitempty"`
	// 迁移任务类型,包括备份文件导入和在线迁移两种类型。
	MigrationType *StopMigrationTaskResponseMigrationType `json:"migration_type,omitempty"`
	// 迁移方式，包括全量迁移和增量迁移两种类型。
	MigrationMethod *StopMigrationTaskResponseMigrationMethod `json:"migration_method,omitempty"`
	BackupFiles     *BackupFilesBody                          `json:"backup_files,omitempty"`
	// 网络类型，包括vpc和vpn两种类型。
	NetworkType    *StopMigrationTaskResponseNetworkType `json:"network_type,omitempty"`
	SourceInstance *SourceInstanceBody                   `json:"source_instance,omitempty"`
	TargetInstance *TargetInstanceBody                   `json:"target_instance,omitempty"`
	// 迁移任务创建时间。
	CreatedAt *string `json:"created_at,omitempty"`
	// 迁移任务完成时间。
	UpdatedAt      *string `json:"updated_at,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o StopMigrationTaskResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "StopMigrationTaskResponse struct{}"
	}

	return strings.Join([]string{"StopMigrationTaskResponse", string(data)}, " ")
}

type StopMigrationTaskResponseStatus struct {
	value string
}

type StopMigrationTaskResponseStatusEnum struct {
	SUCCESS    StopMigrationTaskResponseStatus
	FAILED     StopMigrationTaskResponseStatus
	MIGRATING  StopMigrationTaskResponseStatus
	TERMINATED StopMigrationTaskResponseStatus
}

func GetStopMigrationTaskResponseStatusEnum() StopMigrationTaskResponseStatusEnum {
	return StopMigrationTaskResponseStatusEnum{
		SUCCESS: StopMigrationTaskResponseStatus{
			value: "SUCCESS",
		},
		FAILED: StopMigrationTaskResponseStatus{
			value: "FAILED",
		},
		MIGRATING: StopMigrationTaskResponseStatus{
			value: "MIGRATING",
		},
		TERMINATED: StopMigrationTaskResponseStatus{
			value: "TERMINATED",
		},
	}
}

func (c StopMigrationTaskResponseStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *StopMigrationTaskResponseStatus) UnmarshalJSON(b []byte) error {
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

type StopMigrationTaskResponseMigrationType struct {
	value string
}

type StopMigrationTaskResponseMigrationTypeEnum struct {
	BACKUPFILE_IMPORT StopMigrationTaskResponseMigrationType
	ONLINE_MIGRATION  StopMigrationTaskResponseMigrationType
}

func GetStopMigrationTaskResponseMigrationTypeEnum() StopMigrationTaskResponseMigrationTypeEnum {
	return StopMigrationTaskResponseMigrationTypeEnum{
		BACKUPFILE_IMPORT: StopMigrationTaskResponseMigrationType{
			value: "backupfile_import",
		},
		ONLINE_MIGRATION: StopMigrationTaskResponseMigrationType{
			value: "online_migration",
		},
	}
}

func (c StopMigrationTaskResponseMigrationType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *StopMigrationTaskResponseMigrationType) UnmarshalJSON(b []byte) error {
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

type StopMigrationTaskResponseMigrationMethod struct {
	value string
}

type StopMigrationTaskResponseMigrationMethodEnum struct {
	FULL_AMOUNT_MIGRATION StopMigrationTaskResponseMigrationMethod
	INCREMENTAL_MIGRATION StopMigrationTaskResponseMigrationMethod
}

func GetStopMigrationTaskResponseMigrationMethodEnum() StopMigrationTaskResponseMigrationMethodEnum {
	return StopMigrationTaskResponseMigrationMethodEnum{
		FULL_AMOUNT_MIGRATION: StopMigrationTaskResponseMigrationMethod{
			value: "full_amount_migration",
		},
		INCREMENTAL_MIGRATION: StopMigrationTaskResponseMigrationMethod{
			value: "incremental_migration",
		},
	}
}

func (c StopMigrationTaskResponseMigrationMethod) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *StopMigrationTaskResponseMigrationMethod) UnmarshalJSON(b []byte) error {
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

type StopMigrationTaskResponseNetworkType struct {
	value string
}

type StopMigrationTaskResponseNetworkTypeEnum struct {
	VPC StopMigrationTaskResponseNetworkType
	VPN StopMigrationTaskResponseNetworkType
}

func GetStopMigrationTaskResponseNetworkTypeEnum() StopMigrationTaskResponseNetworkTypeEnum {
	return StopMigrationTaskResponseNetworkTypeEnum{
		VPC: StopMigrationTaskResponseNetworkType{
			value: "vpc",
		},
		VPN: StopMigrationTaskResponseNetworkType{
			value: "vpn",
		},
	}
}

func (c StopMigrationTaskResponseNetworkType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *StopMigrationTaskResponseNetworkType) UnmarshalJSON(b []byte) error {
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

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

// 创建数据迁移任务结构体
type CreateMigrationTaskBody struct {
	// 迁移任务名称。
	TaskName string `json:"task_name"`
	// 迁移任务描述。
	Description *string `json:"description,omitempty"`
	// 迁移任务类型,包括备份文件导入和在线迁移两种类型。 取值范围： - backupfile_import：表示备份文件导入 - online_migration：表示在线迁移。
	MigrationType CreateMigrationTaskBodyMigrationType `json:"migration_type"`
	// 迁移方式，包括全量迁移和增量迁移两种类型。 取值范围： - full_amount_migration：表示全量迁移。 - incremental_migration：表示增量迁移。
	MigrationMethod CreateMigrationTaskBodyMigrationMethod `json:"migration_method"`
	BackupFiles     *BackupFilesBody                       `json:"backup_files,omitempty"`
	// 迁移任务类型为在线迁移时，表示源Redis和目标Redis联通的网络类型，包括vpc和vpn两种类型。
	NetworkType    *CreateMigrationTaskBodyNetworkType `json:"network_type,omitempty"`
	SourceInstance *SourceInstanceBody                 `json:"source_instance,omitempty"`
	TargetInstance *TargetInstanceBody                 `json:"target_instance"`
}

func (o CreateMigrationTaskBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateMigrationTaskBody struct{}"
	}

	return strings.Join([]string{"CreateMigrationTaskBody", string(data)}, " ")
}

type CreateMigrationTaskBodyMigrationType struct {
	value string
}

type CreateMigrationTaskBodyMigrationTypeEnum struct {
	BACKUPFILE_IMPORT CreateMigrationTaskBodyMigrationType
	ONLINE_MIGRATION  CreateMigrationTaskBodyMigrationType
}

func GetCreateMigrationTaskBodyMigrationTypeEnum() CreateMigrationTaskBodyMigrationTypeEnum {
	return CreateMigrationTaskBodyMigrationTypeEnum{
		BACKUPFILE_IMPORT: CreateMigrationTaskBodyMigrationType{
			value: "backupfile_import",
		},
		ONLINE_MIGRATION: CreateMigrationTaskBodyMigrationType{
			value: "online_migration",
		},
	}
}

func (c CreateMigrationTaskBodyMigrationType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateMigrationTaskBodyMigrationType) UnmarshalJSON(b []byte) error {
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

type CreateMigrationTaskBodyMigrationMethod struct {
	value string
}

type CreateMigrationTaskBodyMigrationMethodEnum struct {
	FULL_AMOUNT_MIGRATION CreateMigrationTaskBodyMigrationMethod
	INCREMENTAL_MIGRATION CreateMigrationTaskBodyMigrationMethod
}

func GetCreateMigrationTaskBodyMigrationMethodEnum() CreateMigrationTaskBodyMigrationMethodEnum {
	return CreateMigrationTaskBodyMigrationMethodEnum{
		FULL_AMOUNT_MIGRATION: CreateMigrationTaskBodyMigrationMethod{
			value: "full_amount_migration",
		},
		INCREMENTAL_MIGRATION: CreateMigrationTaskBodyMigrationMethod{
			value: "incremental_migration",
		},
	}
}

func (c CreateMigrationTaskBodyMigrationMethod) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateMigrationTaskBodyMigrationMethod) UnmarshalJSON(b []byte) error {
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

type CreateMigrationTaskBodyNetworkType struct {
	value string
}

type CreateMigrationTaskBodyNetworkTypeEnum struct {
	VPC CreateMigrationTaskBodyNetworkType
	VPN CreateMigrationTaskBodyNetworkType
}

func GetCreateMigrationTaskBodyNetworkTypeEnum() CreateMigrationTaskBodyNetworkTypeEnum {
	return CreateMigrationTaskBodyNetworkTypeEnum{
		VPC: CreateMigrationTaskBodyNetworkType{
			value: "vpc",
		},
		VPN: CreateMigrationTaskBodyNetworkType{
			value: "vpn",
		},
	}
}

func (c CreateMigrationTaskBodyNetworkType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateMigrationTaskBodyNetworkType) UnmarshalJSON(b []byte) error {
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

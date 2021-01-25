/*
 * DDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// 备份信息。
type BackupForList struct {
	// 备份ID。
	Id string `json:"id"`
	// 备份名称。
	Name string `json:"name"`
	// 备份所属的实例ID。
	InstanceId string `json:"instance_id"`
	// 备份所属的实例名称。
	InstanceName string          `json:"instance_name"`
	Datastore    *BackupDatabase `json:"datastore"`
	// 备份类型。 - 取值为“Auto”，表示自动全量备份。 - 取值为“Manual”，表示手动全量备份。 - 取值为“Incremental”，表示自动增量备份。
	Type BackupForListType `json:"type"`
	// 备份开始时间，格式为“yyyy-mm-dd hh:mm:ss”。该时间为UTC时间。
	BeginTime string `json:"begin_time"`
	// 备份结束时间，格式为“yyyy-mm-dd hh:mm:ss”。该时间为UTC时间。
	EndTime string `json:"end_time"`
	// 备份状态。 取值： - BUILDING：备份中。 - COMPLETED：备份完成。 - FAILED：备份失败。 - DISABLED：备份删除中。
	Status BackupForListStatus `json:"status"`
	// 备份大小，单位：KB。
	Size int64 `json:"size"`
	// 备份描述。
	Description string `json:"description"`
}

func (o BackupForList) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BackupForList struct{}"
	}

	return strings.Join([]string{"BackupForList", string(data)}, " ")
}

type BackupForListType struct {
	value string
}

type BackupForListTypeEnum struct {
	AUTO        BackupForListType
	MANUAL      BackupForListType
	FRAGMENT    BackupForListType
	INCREMENTAL BackupForListType
}

func GetBackupForListTypeEnum() BackupForListTypeEnum {
	return BackupForListTypeEnum{
		AUTO: BackupForListType{
			value: "auto",
		},
		MANUAL: BackupForListType{
			value: "manual",
		},
		FRAGMENT: BackupForListType{
			value: "fragment",
		},
		INCREMENTAL: BackupForListType{
			value: "incremental",
		},
	}
}

func (c BackupForListType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BackupForListType) UnmarshalJSON(b []byte) error {
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

type BackupForListStatus struct {
	value string
}

type BackupForListStatusEnum struct {
	BUILDING  BackupForListStatus
	COMPLETED BackupForListStatus
	FAILED    BackupForListStatus
	DELETING  BackupForListStatus
}

func GetBackupForListStatusEnum() BackupForListStatusEnum {
	return BackupForListStatusEnum{
		BUILDING: BackupForListStatus{
			value: "BUILDING",
		},
		COMPLETED: BackupForListStatus{
			value: "COMPLETED",
		},
		FAILED: BackupForListStatus{
			value: "FAILED",
		},
		DELETING: BackupForListStatus{
			value: "DELETING",
		},
	}
}

func (c BackupForListStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BackupForListStatus) UnmarshalJSON(b []byte) error {
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

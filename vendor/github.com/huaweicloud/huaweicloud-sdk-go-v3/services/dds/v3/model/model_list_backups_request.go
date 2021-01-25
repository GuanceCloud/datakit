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

// Request Object
type ListBackupsRequest struct {
	InstanceId *string                       `json:"instance_id,omitempty"`
	BackupId   *string                       `json:"backup_id,omitempty"`
	BackupType *ListBackupsRequestBackupType `json:"backup_type,omitempty"`
	Offset     *int32                        `json:"offset,omitempty"`
	Limit      *int32                        `json:"limit,omitempty"`
	BeginTime  *string                       `json:"begin_time,omitempty"`
	EndTime    *string                       `json:"end_time,omitempty"`
	Mode       *ListBackupsRequestMode       `json:"mode,omitempty"`
}

func (o ListBackupsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListBackupsRequest struct{}"
	}

	return strings.Join([]string{"ListBackupsRequest", string(data)}, " ")
}

type ListBackupsRequestBackupType struct {
	value string
}

type ListBackupsRequestBackupTypeEnum struct {
	AUTO        ListBackupsRequestBackupType
	MANUAL      ListBackupsRequestBackupType
	INCREMENTAL ListBackupsRequestBackupType
}

func GetListBackupsRequestBackupTypeEnum() ListBackupsRequestBackupTypeEnum {
	return ListBackupsRequestBackupTypeEnum{
		AUTO: ListBackupsRequestBackupType{
			value: "Auto",
		},
		MANUAL: ListBackupsRequestBackupType{
			value: "Manual",
		},
		INCREMENTAL: ListBackupsRequestBackupType{
			value: "Incremental",
		},
	}
}

func (c ListBackupsRequestBackupType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListBackupsRequestBackupType) UnmarshalJSON(b []byte) error {
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

type ListBackupsRequestMode struct {
	value string
}

type ListBackupsRequestModeEnum struct {
	SHARDING    ListBackupsRequestMode
	REPLICA_SET ListBackupsRequestMode
	SINGLE      ListBackupsRequestMode
}

func GetListBackupsRequestModeEnum() ListBackupsRequestModeEnum {
	return ListBackupsRequestModeEnum{
		SHARDING: ListBackupsRequestMode{
			value: "Sharding",
		},
		REPLICA_SET: ListBackupsRequestMode{
			value: "ReplicaSet",
		},
		SINGLE: ListBackupsRequestMode{
			value: "Single",
		},
	}
}

func (c ListBackupsRequestMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListBackupsRequestMode) UnmarshalJSON(b []byte) error {
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

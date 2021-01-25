/*
 * RDS
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
	XLanguage  *string                       `json:"X-Language,omitempty"`
	InstanceId string                        `json:"instance_id"`
	BackupId   *string                       `json:"backup_id,omitempty"`
	BackupType *ListBackupsRequestBackupType `json:"backup_type,omitempty"`
	Offset     *int32                        `json:"offset,omitempty"`
	Limit      *int32                        `json:"limit,omitempty"`
	BeginTime  *string                       `json:"begin_time,omitempty"`
	EndTime    *string                       `json:"end_time,omitempty"`
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
	FRAGMENT    ListBackupsRequestBackupType
	INCREMENTAL ListBackupsRequestBackupType
}

func GetListBackupsRequestBackupTypeEnum() ListBackupsRequestBackupTypeEnum {
	return ListBackupsRequestBackupTypeEnum{
		AUTO: ListBackupsRequestBackupType{
			value: "auto",
		},
		MANUAL: ListBackupsRequestBackupType{
			value: "manual",
		},
		FRAGMENT: ListBackupsRequestBackupType{
			value: "fragment",
		},
		INCREMENTAL: ListBackupsRequestBackupType{
			value: "incremental",
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

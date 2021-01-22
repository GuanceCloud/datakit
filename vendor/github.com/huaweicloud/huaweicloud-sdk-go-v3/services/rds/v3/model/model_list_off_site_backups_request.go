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
type ListOffSiteBackupsRequest struct {
	XLanguage  *string                              `json:"X-Language,omitempty"`
	InstanceId string                               `json:"instance_id"`
	BackupId   *string                              `json:"backup_id,omitempty"`
	BackupType *ListOffSiteBackupsRequestBackupType `json:"backup_type,omitempty"`
	Offset     *int32                               `json:"offset,omitempty"`
	Limit      *int32                               `json:"limit,omitempty"`
	BeginTime  *string                              `json:"begin_time,omitempty"`
	EndTime    *string                              `json:"end_time,omitempty"`
}

func (o ListOffSiteBackupsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListOffSiteBackupsRequest struct{}"
	}

	return strings.Join([]string{"ListOffSiteBackupsRequest", string(data)}, " ")
}

type ListOffSiteBackupsRequestBackupType struct {
	value string
}

type ListOffSiteBackupsRequestBackupTypeEnum struct {
	AUTO        ListOffSiteBackupsRequestBackupType
	INCREMENTAL ListOffSiteBackupsRequestBackupType
}

func GetListOffSiteBackupsRequestBackupTypeEnum() ListOffSiteBackupsRequestBackupTypeEnum {
	return ListOffSiteBackupsRequestBackupTypeEnum{
		AUTO: ListOffSiteBackupsRequestBackupType{
			value: "auto",
		},
		INCREMENTAL: ListOffSiteBackupsRequestBackupType{
			value: "incremental",
		},
	}
}

func (c ListOffSiteBackupsRequestBackupType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ListOffSiteBackupsRequestBackupType) UnmarshalJSON(b []byte) error {
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

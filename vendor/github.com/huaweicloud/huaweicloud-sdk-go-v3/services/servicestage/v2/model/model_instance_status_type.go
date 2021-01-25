/*
 * ServiceStage
 *
 * ServiceStage的API,包括应用管理和仓库授权管理
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

// 实例状态。
type InstanceStatusType struct {
	value string
}

type InstanceStatusTypeEnum struct {
	INITIALIZING     InstanceStatusType
	UPGRADING        InstanceStatusType
	FAILED           InstanceStatusType
	RUNNING          InstanceStatusType
	DOWN             InstanceStatusType
	DELETING         InstanceStatusType
	DELETED          InstanceStatusType
	RESERVED         InstanceStatusType
	STARTING         InstanceStatusType
	STOPPING         InstanceStatusType
	STOPPED          InstanceStatusType
	RESTARTING       InstanceStatusType
	PENDING          InstanceStatusType
	UNKNOWN          InstanceStatusType
	PARTIALLY_FAILED InstanceStatusType
}

func GetInstanceStatusTypeEnum() InstanceStatusTypeEnum {
	return InstanceStatusTypeEnum{
		INITIALIZING: InstanceStatusType{
			value: "INITIALIZING",
		},
		UPGRADING: InstanceStatusType{
			value: "UPGRADING",
		},
		FAILED: InstanceStatusType{
			value: "FAILED",
		},
		RUNNING: InstanceStatusType{
			value: "RUNNING",
		},
		DOWN: InstanceStatusType{
			value: "DOWN",
		},
		DELETING: InstanceStatusType{
			value: "DELETING",
		},
		DELETED: InstanceStatusType{
			value: "DELETED",
		},
		RESERVED: InstanceStatusType{
			value: "RESERVED",
		},
		STARTING: InstanceStatusType{
			value: "STARTING",
		},
		STOPPING: InstanceStatusType{
			value: "STOPPING",
		},
		STOPPED: InstanceStatusType{
			value: "STOPPED",
		},
		RESTARTING: InstanceStatusType{
			value: "RESTARTING",
		},
		PENDING: InstanceStatusType{
			value: "PENDING",
		},
		UNKNOWN: InstanceStatusType{
			value: "UNKNOWN",
		},
		PARTIALLY_FAILED: InstanceStatusType{
			value: "PARTIALLY_FAILED",
		},
	}
}

func (c InstanceStatusType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *InstanceStatusType) UnmarshalJSON(b []byte) error {
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

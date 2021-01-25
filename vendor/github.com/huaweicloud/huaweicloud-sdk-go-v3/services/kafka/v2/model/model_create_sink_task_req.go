/*
 * Kafka
 *
 * Kafka Document API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

type CreateSinkTaskReq struct {
	// 源数据类型，目前只支持BLOB。
	SourceType CreateSinkTaskReqSourceType `json:"source_type"`
	// 转储任务名称。
	TaskName string `json:"task_name"`
	// 转存的目标类型，当前只支持OBS。
	DestinationType          CreateSinkTaskReqDestinationType `json:"destination_type"`
	ObsDestinationDescriptor *ObsDestinationDescriptor        `json:"obs_destination_descriptor"`
}

func (o CreateSinkTaskReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateSinkTaskReq struct{}"
	}

	return strings.Join([]string{"CreateSinkTaskReq", string(data)}, " ")
}

type CreateSinkTaskReqSourceType struct {
	value string
}

type CreateSinkTaskReqSourceTypeEnum struct {
	BLOB CreateSinkTaskReqSourceType
}

func GetCreateSinkTaskReqSourceTypeEnum() CreateSinkTaskReqSourceTypeEnum {
	return CreateSinkTaskReqSourceTypeEnum{
		BLOB: CreateSinkTaskReqSourceType{
			value: "BLOB",
		},
	}
}

func (c CreateSinkTaskReqSourceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateSinkTaskReqSourceType) UnmarshalJSON(b []byte) error {
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

type CreateSinkTaskReqDestinationType struct {
	value string
}

type CreateSinkTaskReqDestinationTypeEnum struct {
	OBS CreateSinkTaskReqDestinationType
}

func GetCreateSinkTaskReqDestinationTypeEnum() CreateSinkTaskReqDestinationTypeEnum {
	return CreateSinkTaskReqDestinationTypeEnum{
		OBS: CreateSinkTaskReqDestinationType{
			value: "OBS",
		},
	}
}

func (c CreateSinkTaskReqDestinationType) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateSinkTaskReqDestinationType) UnmarshalJSON(b []byte) error {
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

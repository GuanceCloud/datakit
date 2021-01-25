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

type BatchRestartOrDeleteInstanceReq struct {
	// 实例的ID列表。
	Instances *[]string `json:"instances,omitempty"`
	// 对实例的操作：restart、delete
	Action BatchRestartOrDeleteInstanceReqAction `json:"action"`
	// 参数值为kafka，表示删除租户所有创建失败的Kafka实例。
	AllFailure *BatchRestartOrDeleteInstanceReqAllFailure `json:"all_failure,omitempty"`
}

func (o BatchRestartOrDeleteInstanceReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchRestartOrDeleteInstanceReq struct{}"
	}

	return strings.Join([]string{"BatchRestartOrDeleteInstanceReq", string(data)}, " ")
}

type BatchRestartOrDeleteInstanceReqAction struct {
	value string
}

type BatchRestartOrDeleteInstanceReqActionEnum struct {
	RESTART BatchRestartOrDeleteInstanceReqAction
	DELETE  BatchRestartOrDeleteInstanceReqAction
}

func GetBatchRestartOrDeleteInstanceReqActionEnum() BatchRestartOrDeleteInstanceReqActionEnum {
	return BatchRestartOrDeleteInstanceReqActionEnum{
		RESTART: BatchRestartOrDeleteInstanceReqAction{
			value: "restart",
		},
		DELETE: BatchRestartOrDeleteInstanceReqAction{
			value: "delete",
		},
	}
}

func (c BatchRestartOrDeleteInstanceReqAction) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BatchRestartOrDeleteInstanceReqAction) UnmarshalJSON(b []byte) error {
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

type BatchRestartOrDeleteInstanceReqAllFailure struct {
	value string
}

type BatchRestartOrDeleteInstanceReqAllFailureEnum struct {
	TRUE  BatchRestartOrDeleteInstanceReqAllFailure
	FALSE BatchRestartOrDeleteInstanceReqAllFailure
	KAFKA BatchRestartOrDeleteInstanceReqAllFailure
}

func GetBatchRestartOrDeleteInstanceReqAllFailureEnum() BatchRestartOrDeleteInstanceReqAllFailureEnum {
	return BatchRestartOrDeleteInstanceReqAllFailureEnum{
		TRUE: BatchRestartOrDeleteInstanceReqAllFailure{
			value: "true",
		},
		FALSE: BatchRestartOrDeleteInstanceReqAllFailure{
			value: "false",
		},
		KAFKA: BatchRestartOrDeleteInstanceReqAllFailure{
			value: "kafka",
		},
	}
}

func (c BatchRestartOrDeleteInstanceReqAllFailure) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *BatchRestartOrDeleteInstanceReqAllFailure) UnmarshalJSON(b []byte) error {
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

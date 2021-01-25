/*
 * RabbitMQ
 *
 * RabbitMQ Document API
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
	// 是否批量删除创建失败的实例。  当参数值为“true”时，删除租户所有创建失败的实例，此时请求参数instances可为空。
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
}

func GetBatchRestartOrDeleteInstanceReqAllFailureEnum() BatchRestartOrDeleteInstanceReqAllFailureEnum {
	return BatchRestartOrDeleteInstanceReqAllFailureEnum{
		TRUE: BatchRestartOrDeleteInstanceReqAllFailure{
			value: "true",
		},
		FALSE: BatchRestartOrDeleteInstanceReqAllFailure{
			value: "false",
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

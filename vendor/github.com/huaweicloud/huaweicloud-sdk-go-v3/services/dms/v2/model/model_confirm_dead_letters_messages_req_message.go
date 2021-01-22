/*
 * DMS
 *
 * DMS Document API
 *
 */

package model

import (
	"encoding/json"
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"
	"strings"
)

type ConfirmDeadLettersMessagesReqMessage struct {
	// 消费时返回的ID。
	Handler *string `json:"handler,omitempty"`
	// 客户端处理数据的状态。 取值为“success”或者“fail”。
	Status *ConfirmDeadLettersMessagesReqMessageStatus `json:"status,omitempty"`
}

func (o ConfirmDeadLettersMessagesReqMessage) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ConfirmDeadLettersMessagesReqMessage struct{}"
	}

	return strings.Join([]string{"ConfirmDeadLettersMessagesReqMessage", string(data)}, " ")
}

type ConfirmDeadLettersMessagesReqMessageStatus struct {
	value string
}

type ConfirmDeadLettersMessagesReqMessageStatusEnum struct {
	SUCCESS ConfirmDeadLettersMessagesReqMessageStatus
	FAIL    ConfirmDeadLettersMessagesReqMessageStatus
}

func GetConfirmDeadLettersMessagesReqMessageStatusEnum() ConfirmDeadLettersMessagesReqMessageStatusEnum {
	return ConfirmDeadLettersMessagesReqMessageStatusEnum{
		SUCCESS: ConfirmDeadLettersMessagesReqMessageStatus{
			value: "success",
		},
		FAIL: ConfirmDeadLettersMessagesReqMessageStatus{
			value: "fail",
		},
	}
}

func (c ConfirmDeadLettersMessagesReqMessageStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ConfirmDeadLettersMessagesReqMessageStatus) UnmarshalJSON(b []byte) error {
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

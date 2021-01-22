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

type CreateQueueReq struct {
	// 队列的名称，必须唯一。  长度不超过64位的字符串，包含a~z，A~Z，0~9、中划线（-）和下划线（_）。  创建队列后无法修改名称。
	Name string `json:"name"`
	// 队列类型。  取值范围： - NORMAL：普通队列，更高的并发性能，不保证先入先出（FIFO）的严格顺序。 - FIFO：有序队列，保证消息先入先出（FIFO）的严格顺序。 - KAFKA_HA：高可靠模式的kafka队列。消息多副本同步落盘，保证消息的可靠性。 - KAFKA_HT：高吞吐模式的kafka队列。消息副本异步落盘，具有较高的性能。  默认值：NORMAL
	QueueMode *CreateQueueReqQueueMode `json:"queue_mode,omitempty"`
	// 队列的描述信息。  长度不超过160位的字符串，不能包含尖括号<>。
	Description *string `json:"description,omitempty"`
	// 仅当queue_mode为“NORMAL”或者“FIFO”时，该参数有效。  是否开启死信消息，死信消息是指无法被正常消费的消息。  当达到最大消费次数仍然失败后，DMS会将该条消息转存到死信队列中，有效期为72小时，用户可以根据需要对死信消息进行重新消费。  消费死信消息时，只能消费该消费组产生的死信消息。  有序队列的死信消息依然按照先入先出（FIFO）的顺序存储在死信队列中。  取值范围： - enable：开启 - disable：不开启  默认值：disable
	RedrivePolicy *CreateQueueReqRedrivePolicy `json:"redrive_policy,omitempty"`
	// 仅当redrive_policy为enable时，该参数必选。  最大确认消费失败的次数，当达到最大确认失败次数后，DMS会将该条消息转存到死信队列中。  取值范围：1~100
	MaxConsumeCount *int32 `json:"max_consume_count,omitempty"`
	// 指定kafka队列的消息保存时间，单位为小时。  仅当queue_mode为KAFKA_HA或者KAFKA_HT才有效。  取值范围: 1-72（小时）
	RetentionHours *int32 `json:"retention_hours,omitempty"`
}

func (o CreateQueueReq) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateQueueReq struct{}"
	}

	return strings.Join([]string{"CreateQueueReq", string(data)}, " ")
}

type CreateQueueReqQueueMode struct {
	value string
}

type CreateQueueReqQueueModeEnum struct {
	NORMAL   CreateQueueReqQueueMode
	FIFO     CreateQueueReqQueueMode
	KAFKA_HA CreateQueueReqQueueMode
	KAFKA_HT CreateQueueReqQueueMode
}

func GetCreateQueueReqQueueModeEnum() CreateQueueReqQueueModeEnum {
	return CreateQueueReqQueueModeEnum{
		NORMAL: CreateQueueReqQueueMode{
			value: "NORMAL",
		},
		FIFO: CreateQueueReqQueueMode{
			value: "FIFO",
		},
		KAFKA_HA: CreateQueueReqQueueMode{
			value: "KAFKA_HA",
		},
		KAFKA_HT: CreateQueueReqQueueMode{
			value: "KAFKA_HT",
		},
	}
}

func (c CreateQueueReqQueueMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateQueueReqQueueMode) UnmarshalJSON(b []byte) error {
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

type CreateQueueReqRedrivePolicy struct {
	value string
}

type CreateQueueReqRedrivePolicyEnum struct {
	ENABLE  CreateQueueReqRedrivePolicy
	DISABLE CreateQueueReqRedrivePolicy
}

func GetCreateQueueReqRedrivePolicyEnum() CreateQueueReqRedrivePolicyEnum {
	return CreateQueueReqRedrivePolicyEnum{
		ENABLE: CreateQueueReqRedrivePolicy{
			value: "enable",
		},
		DISABLE: CreateQueueReqRedrivePolicy{
			value: "disable",
		},
	}
}

func (c CreateQueueReqRedrivePolicy) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *CreateQueueReqRedrivePolicy) UnmarshalJSON(b []byte) error {
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

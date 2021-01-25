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

// Response Object
type ShowQueueResponse struct {
	// 队列ID。
	Id *string `json:"id,omitempty"`
	// 队列的名称。
	Name *string `json:"name,omitempty"`
	// 队列的描述信息。
	Description *string `json:"description,omitempty"`
	// 队列类型。
	QueueMode *ShowQueueResponseQueueMode `json:"queue_mode,omitempty"`
	// 消息在队列中允许保留的时长（单位分钟）。
	Reservation *int32 `json:"reservation,omitempty"`
	// 队列中允许的最大消息大小（单位Byte）。
	MaxMsgSizeByte *int32 `json:"max_msg_size_byte,omitempty"`
	// 队列的消息总数。
	ProducedMessages *int32 `json:"produced_messages,omitempty"`
	// 该队列是否开启死信消息。仅当include_deadletter为true时，才有该响应参数。 - enable：表示开启。 - disable：表示不开启。
	RedrivePolicy *ShowQueueResponseRedrivePolicy `json:"redrive_policy,omitempty"`
	// 最大确认消费失败的次数，当达到最大确认失败次数后，DMS会将该条消息转存到死信队列中。 仅当include_deadletter为true时，才有该响应参数。
	MaxConsumeCount *int32 `json:"max_consume_count,omitempty"`
	// 该队列下的消费组数量。
	GroupCount *int32 `json:"group_count,omitempty"`
	// 仅Kafka队列才有该参数。
	KafkaTopic     *string `json:"kafka_topic,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ShowQueueResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowQueueResponse struct{}"
	}

	return strings.Join([]string{"ShowQueueResponse", string(data)}, " ")
}

type ShowQueueResponseQueueMode struct {
	value string
}

type ShowQueueResponseQueueModeEnum struct {
	NORMAL   ShowQueueResponseQueueMode
	FIFO     ShowQueueResponseQueueMode
	KAFKA_HA ShowQueueResponseQueueMode
	KAFKA_HT ShowQueueResponseQueueMode
}

func GetShowQueueResponseQueueModeEnum() ShowQueueResponseQueueModeEnum {
	return ShowQueueResponseQueueModeEnum{
		NORMAL: ShowQueueResponseQueueMode{
			value: "NORMAL",
		},
		FIFO: ShowQueueResponseQueueMode{
			value: "FIFO",
		},
		KAFKA_HA: ShowQueueResponseQueueMode{
			value: "KAFKA_HA",
		},
		KAFKA_HT: ShowQueueResponseQueueMode{
			value: "KAFKA_HT",
		},
	}
}

func (c ShowQueueResponseQueueMode) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowQueueResponseQueueMode) UnmarshalJSON(b []byte) error {
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

type ShowQueueResponseRedrivePolicy struct {
	value string
}

type ShowQueueResponseRedrivePolicyEnum struct {
	ENABLE  ShowQueueResponseRedrivePolicy
	DISABLE ShowQueueResponseRedrivePolicy
}

func GetShowQueueResponseRedrivePolicyEnum() ShowQueueResponseRedrivePolicyEnum {
	return ShowQueueResponseRedrivePolicyEnum{
		ENABLE: ShowQueueResponseRedrivePolicy{
			value: "enable",
		},
		DISABLE: ShowQueueResponseRedrivePolicy{
			value: "disable",
		},
	}
}

func (c ShowQueueResponseRedrivePolicy) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

func (c *ShowQueueResponseRedrivePolicy) UnmarshalJSON(b []byte) error {
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

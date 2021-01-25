/*
 * DMS
 *
 * DMS Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

type ListQueuesRespQueues struct {
	// 队列ID。
	Id *string `json:"id,omitempty"`
	// 队列的名称。
	Name *string `json:"name,omitempty"`
	// 队列的描述信息。
	Description *string `json:"description,omitempty"`
	// 队列类型。
	QueueMode *string `json:"queue_mode,omitempty"`
	// 消息在队列中允许保留的时长（单位分钟）。
	Reservation *int32 `json:"reservation,omitempty"`
	// 队列中允许的最大消息大小（单位Byte）。
	MaxMsgSizeByte *int32 `json:"max_msg_size_byte,omitempty"`
	// 队列的消息总数。
	ProducedMessages *int32 `json:"produced_messages,omitempty"`
	// 该队列是否开启死信消息。仅当include_deadletter为true时，才有该响应参数。 - enable：表示开启。 - disable：表示不开启。
	RedrivePolicy *string `json:"redrive_policy,omitempty"`
	// 最大确认消费失败的次数，当达到最大确认失败次数后，DMS会将该条消息转存到死信队列中。  仅当include_deadletter为true时，才有该响应参数。
	MaxConsumeCount *int32 `json:"max_consume_count,omitempty"`
	// 该队列下的消费组数量。
	GroupCount *int32 `json:"group_count,omitempty"`
}

func (o ListQueuesRespQueues) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListQueuesRespQueues struct{}"
	}

	return strings.Join([]string{"ListQueuesRespQueues", string(data)}, " ")
}

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

// 消费组信息
type ListQueueGroupsRespGroups struct {
	// 队列的名称。
	Id *string `json:"id,omitempty"`
	// 队列的名称。
	Name *string `json:"name,omitempty"`
	// 队列的消息总数，不包含过期删除的消息数。
	ProducedMessages *int32 `json:"produced_messages,omitempty"`
	// 已正常消费的消息总数。
	ConsumedMessages *int32 `json:"consumed_messages,omitempty"`
	// 该消费组可以消费的普通消息数。
	AvailableMessages *int32 `json:"available_messages,omitempty"`
	// 该消费组产生的死信息消息总数。仅当include_deadletter为true时，才有该响应参数。
	ProducedDeadletters *int32 `json:"produced_deadletters,omitempty"`
	// 该消费组未消费的死信消息数。仅当include_deadletter为true时，才有该响应参数。
	AvailableDeadletters *int32 `json:"available_deadletters,omitempty"`
}

func (o ListQueueGroupsRespGroups) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListQueueGroupsRespGroups struct{}"
	}

	return strings.Join([]string{"ListQueueGroupsRespGroups", string(data)}, " ")
}

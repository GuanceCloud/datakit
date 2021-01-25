/*
 * Kafka
 *
 * Kafka Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// 消费组信息。
type ShowGroupsRespGroup struct {
	// 消费组名称。
	GroupId *string `json:"group_id,omitempty"`
	// 消费组状态。包含以下状态： - Dead：消费组内没有任何成员，且没有任何元数据。 - Empty：消费组内没有任何成员，存在元数据。 - PreparingRebalance：准备开启rebalance。 - CompletingRebalance：所有成员加入group。 - Stable：消费组内成员可正常消费。
	State *string `json:"state,omitempty"`
	// 协调器编号。
	CoordinatorId *int32 `json:"coordinator_id,omitempty"`
	// 消费者列表。
	Members *[]ShowGroupsRespGroupMembers `json:"members,omitempty"`
	// 消费进度。
	GroupMessageOffsets *[]ShowGroupsRespGroupGroupMessageOffsets `json:"group_message_offsets,omitempty"`
	// 分区分配策略。
	AssignmentStrategy *string `json:"assignment_strategy,omitempty"`
}

func (o ShowGroupsRespGroup) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowGroupsRespGroup struct{}"
	}

	return strings.Join([]string{"ShowGroupsRespGroup", string(data)}, " ")
}

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

type ShowGroupsRespGroupMembers struct {
	// 消费组consumer地址。
	Host *string `json:"host,omitempty"`
	// consumer分配到的分区信息。
	Assignment *[]ShowGroupsRespGroupAssignment `json:"assignment,omitempty"`
	// 消费组consumer的ID。
	MemberId *string `json:"member_id,omitempty"`
	// 客户端ID。
	ClientId *string `json:"client_id,omitempty"`
}

func (o ShowGroupsRespGroupMembers) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowGroupsRespGroupMembers struct{}"
	}

	return strings.Join([]string{"ShowGroupsRespGroupMembers", string(data)}, " ")
}

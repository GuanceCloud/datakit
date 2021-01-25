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

// Request Object
type ConsumeDeadlettersMessageRequest struct {
	ProjectId       string `json:"project_id"`
	QueueId         string `json:"queue_id"`
	ConsumerGroupId string `json:"consumer_group_id"`
	MaxMsgs         *int32 `json:"max_msgs,omitempty"`
	TimeWait        *int32 `json:"time_wait,omitempty"`
	AckWait         *int32 `json:"ack_wait,omitempty"`
}

func (o ConsumeDeadlettersMessageRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ConsumeDeadlettersMessageRequest struct{}"
	}

	return strings.Join([]string{"ConsumeDeadlettersMessageRequest", string(data)}, " ")
}

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
type ConfirmDeadLettersMessagesRequest struct {
	ProjectId       string                         `json:"project_id"`
	QueueId         string                         `json:"queue_id"`
	ConsumerGroupId string                         `json:"consumer_group_id"`
	Body            *ConfirmDeadLettersMessagesReq `json:"body,omitempty"`
}

func (o ConfirmDeadLettersMessagesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ConfirmDeadLettersMessagesRequest struct{}"
	}

	return strings.Join([]string{"ConfirmDeadLettersMessagesRequest", string(data)}, " ")
}

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
type CreateConsumerGroupRequest struct {
	ProjectId string                  `json:"project_id"`
	QueueId   string                  `json:"queue_id"`
	Body      *CreateConsumerGroupReq `json:"body,omitempty"`
}

func (o CreateConsumerGroupRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateConsumerGroupRequest struct{}"
	}

	return strings.Join([]string{"CreateConsumerGroupRequest", string(data)}, " ")
}

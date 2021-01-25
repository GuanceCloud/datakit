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
type ShowQueueTagsRequest struct {
	ProjectId string `json:"project_id"`
	QueueId   string `json:"queue_id"`
}

func (o ShowQueueTagsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowQueueTagsRequest struct{}"
	}

	return strings.Join([]string{"ShowQueueTagsRequest", string(data)}, " ")
}

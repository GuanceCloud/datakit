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

// Request Object
type ListSinkTasksRequest struct {
	ProjectId   string `json:"project_id"`
	ConnectorId string `json:"connector_id"`
}

func (o ListSinkTasksRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListSinkTasksRequest struct{}"
	}

	return strings.Join([]string{"ListSinkTasksRequest", string(data)}, " ")
}

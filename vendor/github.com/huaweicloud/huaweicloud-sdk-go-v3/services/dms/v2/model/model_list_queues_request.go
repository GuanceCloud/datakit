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
type ListQueuesRequest struct {
	ProjectId         string `json:"project_id"`
	IncludeDeadletter *bool  `json:"include_deadletter,omitempty"`
}

func (o ListQueuesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListQueuesRequest struct{}"
	}

	return strings.Join([]string{"ListQueuesRequest", string(data)}, " ")
}

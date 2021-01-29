/*
 * ECS
 *
 * ECS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type BatchStopServersRequest struct {
	Body *BatchStopServersRequestBody `json:"body,omitempty"`
}

func (o BatchStopServersRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchStopServersRequest struct{}"
	}

	return strings.Join([]string{"BatchStopServersRequest", string(data)}, " ")
}

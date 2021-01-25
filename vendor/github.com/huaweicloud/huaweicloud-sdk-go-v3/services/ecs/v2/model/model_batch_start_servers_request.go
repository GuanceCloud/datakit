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
type BatchStartServersRequest struct {
	Body *BatchStartServersRequestBody `json:"body,omitempty"`
}

func (o BatchStartServersRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchStartServersRequest struct{}"
	}

	return strings.Join([]string{"BatchStartServersRequest", string(data)}, " ")
}

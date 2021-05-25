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
type ResizeServerRequest struct {
	ServerId string                   `json:"server_id"`
	Body     *ResizeServerRequestBody `json:"body,omitempty"`
}

func (o ResizeServerRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResizeServerRequest struct{}"
	}

	return strings.Join([]string{"ResizeServerRequest", string(data)}, " ")
}

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
type ResizePostPaidServerRequest struct {
	ServerId string                           `json:"server_id"`
	Body     *ResizePostPaidServerRequestBody `json:"body,omitempty"`
}

func (o ResizePostPaidServerRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResizePostPaidServerRequest struct{}"
	}

	return strings.Join([]string{"ResizePostPaidServerRequest", string(data)}, " ")
}

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
type ResetServerPasswordRequest struct {
	ServerId string                          `json:"server_id"`
	Body     *ResetServerPasswordRequestBody `json:"body,omitempty"`
}

func (o ResetServerPasswordRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResetServerPasswordRequest struct{}"
	}

	return strings.Join([]string{"ResetServerPasswordRequest", string(data)}, " ")
}

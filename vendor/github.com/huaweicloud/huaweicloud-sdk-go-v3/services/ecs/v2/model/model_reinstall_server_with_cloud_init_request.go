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
type ReinstallServerWithCloudInitRequest struct {
	ServerId string                                   `json:"server_id"`
	Body     *ReinstallServerWithCloudInitRequestBody `json:"body,omitempty"`
}

func (o ReinstallServerWithCloudInitRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ReinstallServerWithCloudInitRequest struct{}"
	}

	return strings.Join([]string{"ReinstallServerWithCloudInitRequest", string(data)}, " ")
}

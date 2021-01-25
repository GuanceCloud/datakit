/*
 * RabbitMQ
 *
 * RabbitMQ Document API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ShowMaintainWindowsRequest struct {
}

func (o ShowMaintainWindowsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowMaintainWindowsRequest struct{}"
	}

	return strings.Join([]string{"ShowMaintainWindowsRequest", string(data)}, " ")
}

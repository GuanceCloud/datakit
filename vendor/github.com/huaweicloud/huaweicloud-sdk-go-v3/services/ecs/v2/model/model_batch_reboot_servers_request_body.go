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

// This is a auto create Body Object
type BatchRebootServersRequestBody struct {
	Reboot *BatchRebootSeversOption `json:"reboot"`
}

func (o BatchRebootServersRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchRebootServersRequestBody struct{}"
	}

	return strings.Join([]string{"BatchRebootServersRequestBody", string(data)}, " ")
}

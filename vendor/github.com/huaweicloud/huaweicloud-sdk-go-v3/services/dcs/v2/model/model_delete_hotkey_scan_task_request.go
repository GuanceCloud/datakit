/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type DeleteHotkeyScanTaskRequest struct {
	InstanceId string `json:"instance_id"`
	HotkeyId   string `json:"hotkey_id"`
}

func (o DeleteHotkeyScanTaskRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteHotkeyScanTaskRequest struct{}"
	}

	return strings.Join([]string{"DeleteHotkeyScanTaskRequest", string(data)}, " ")
}

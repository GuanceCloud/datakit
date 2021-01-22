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
type ShowHotkeyTaskDetailsRequest struct {
	InstanceId string `json:"instance_id"`
	HotkeyId   string `json:"hotkey_id"`
}

func (o ShowHotkeyTaskDetailsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowHotkeyTaskDetailsRequest struct{}"
	}

	return strings.Join([]string{"ShowHotkeyTaskDetailsRequest", string(data)}, " ")
}

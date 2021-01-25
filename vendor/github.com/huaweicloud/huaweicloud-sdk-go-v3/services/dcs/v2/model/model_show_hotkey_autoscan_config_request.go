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
type ShowHotkeyAutoscanConfigRequest struct {
	InstanceId string `json:"instance_id"`
}

func (o ShowHotkeyAutoscanConfigRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowHotkeyAutoscanConfigRequest struct{}"
	}

	return strings.Join([]string{"ShowHotkeyAutoscanConfigRequest", string(data)}, " ")
}

/*
 * BMS
 *
 * BMS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ShowWindowsBaremetalServerPwdRequest struct {
	ServerId string `json:"server_id"`
}

func (o ShowWindowsBaremetalServerPwdRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowWindowsBaremetalServerPwdRequest struct{}"
	}

	return strings.Join([]string{"ShowWindowsBaremetalServerPwdRequest", string(data)}, " ")
}

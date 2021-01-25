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
type ShowResetPwdRequest struct {
	ServerId string `json:"server_id"`
}

func (o ShowResetPwdRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowResetPwdRequest struct{}"
	}

	return strings.Join([]string{"ShowResetPwdRequest", string(data)}, " ")
}

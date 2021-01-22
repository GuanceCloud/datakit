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
type WindowsBaremetalServerCleanPwdRequest struct {
	ServerId string `json:"server_id"`
}

func (o WindowsBaremetalServerCleanPwdRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "WindowsBaremetalServerCleanPwdRequest struct{}"
	}

	return strings.Join([]string{"WindowsBaremetalServerCleanPwdRequest", string(data)}, " ")
}

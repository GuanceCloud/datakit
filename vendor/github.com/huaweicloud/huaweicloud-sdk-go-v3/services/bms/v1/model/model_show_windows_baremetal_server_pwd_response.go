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

// Response Object
type ShowWindowsBaremetalServerPwdResponse struct {
	// 加密后的密码
	Password       *string `json:"password,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ShowWindowsBaremetalServerPwdResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowWindowsBaremetalServerPwdResponse struct{}"
	}

	return strings.Join([]string{"ShowWindowsBaremetalServerPwdResponse", string(data)}, " ")
}

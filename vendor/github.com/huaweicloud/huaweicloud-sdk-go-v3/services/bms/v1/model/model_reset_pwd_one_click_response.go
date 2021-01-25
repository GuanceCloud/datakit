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
type ResetPwdOneClickResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o ResetPwdOneClickResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResetPwdOneClickResponse struct{}"
	}

	return strings.Join([]string{"ResetPwdOneClickResponse", string(data)}, " ")
}

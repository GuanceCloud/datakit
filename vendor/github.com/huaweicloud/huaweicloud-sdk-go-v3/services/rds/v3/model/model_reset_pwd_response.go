/*
 * RDS
 *
 * API v3
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ResetPwdResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o ResetPwdResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResetPwdResponse struct{}"
	}

	return strings.Join([]string{"ResetPwdResponse", string(data)}, " ")
}

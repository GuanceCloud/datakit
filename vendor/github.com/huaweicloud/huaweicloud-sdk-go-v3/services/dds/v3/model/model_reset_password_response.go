/*
 * DDS
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
type ResetPasswordResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o ResetPasswordResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResetPasswordResponse struct{}"
	}

	return strings.Join([]string{"ResetPasswordResponse", string(data)}, " ")
}

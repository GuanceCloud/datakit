/*
 * SWR
 *
 * SWR API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type DeleteUserRepositoryAuthResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteUserRepositoryAuthResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteUserRepositoryAuthResponse struct{}"
	}

	return strings.Join([]string{"DeleteUserRepositoryAuthResponse", string(data)}, " ")
}

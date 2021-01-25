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
type UpdateUserRepositoryAuthResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o UpdateUserRepositoryAuthResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateUserRepositoryAuthResponse struct{}"
	}

	return strings.Join([]string{"UpdateUserRepositoryAuthResponse", string(data)}, " ")
}

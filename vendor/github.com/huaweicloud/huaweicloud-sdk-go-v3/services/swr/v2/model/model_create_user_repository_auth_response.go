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
type CreateUserRepositoryAuthResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o CreateUserRepositoryAuthResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateUserRepositoryAuthResponse struct{}"
	}

	return strings.Join([]string{"CreateUserRepositoryAuthResponse", string(data)}, " ")
}

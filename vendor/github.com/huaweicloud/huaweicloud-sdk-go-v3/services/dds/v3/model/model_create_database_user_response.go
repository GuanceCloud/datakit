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
type CreateDatabaseUserResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o CreateDatabaseUserResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateDatabaseUserResponse struct{}"
	}

	return strings.Join([]string{"CreateDatabaseUserResponse", string(data)}, " ")
}

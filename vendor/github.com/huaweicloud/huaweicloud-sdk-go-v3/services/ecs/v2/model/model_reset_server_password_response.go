/*
 * ECS
 *
 * ECS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ResetServerPasswordResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o ResetServerPasswordResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResetServerPasswordResponse struct{}"
	}

	return strings.Join([]string{"ResetServerPasswordResponse", string(data)}, " ")
}

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

// This is a auto create Body Object
type UpdateServerRequestBody struct {
	Server *UpdateServerOption `json:"server"`
}

func (o UpdateServerRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateServerRequestBody struct{}"
	}

	return strings.Join([]string{"UpdateServerRequestBody", string(data)}, " ")
}

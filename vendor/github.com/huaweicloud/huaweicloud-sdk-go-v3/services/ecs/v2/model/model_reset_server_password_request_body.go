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
type ResetServerPasswordRequestBody struct {
	ResetPassword *ResetServerPasswordOption `json:"reset-password"`
}

func (o ResetServerPasswordRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ResetServerPasswordRequestBody struct{}"
	}

	return strings.Join([]string{"ResetServerPasswordRequestBody", string(data)}, " ")
}

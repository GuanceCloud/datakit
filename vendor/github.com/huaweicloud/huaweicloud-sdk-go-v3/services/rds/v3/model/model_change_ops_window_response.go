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
type ChangeOpsWindowResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o ChangeOpsWindowResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ChangeOpsWindowResponse struct{}"
	}

	return strings.Join([]string{"ChangeOpsWindowResponse", string(data)}, " ")
}

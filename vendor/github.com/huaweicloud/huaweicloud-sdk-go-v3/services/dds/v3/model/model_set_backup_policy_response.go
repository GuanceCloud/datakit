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
type SetBackupPolicyResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o SetBackupPolicyResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "SetBackupPolicyResponse struct{}"
	}

	return strings.Join([]string{"SetBackupPolicyResponse", string(data)}, " ")
}

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
type DeleteManualBackupResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteManualBackupResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteManualBackupResponse struct{}"
	}

	return strings.Join([]string{"DeleteManualBackupResponse", string(data)}, " ")
}

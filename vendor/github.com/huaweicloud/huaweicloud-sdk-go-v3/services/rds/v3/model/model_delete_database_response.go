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
type DeleteDatabaseResponse struct {
	// 操作结果。
	Resp           *string `json:"resp,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o DeleteDatabaseResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteDatabaseResponse struct{}"
	}

	return strings.Join([]string{"DeleteDatabaseResponse", string(data)}, " ")
}

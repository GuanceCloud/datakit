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
type RevokeResponse struct {
	// 操作结果。
	Resp           *string `json:"resp,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o RevokeResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RevokeResponse struct{}"
	}

	return strings.Join([]string{"RevokeResponse", string(data)}, " ")
}

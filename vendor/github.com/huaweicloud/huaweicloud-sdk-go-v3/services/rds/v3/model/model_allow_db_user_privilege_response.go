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
type AllowDbUserPrivilegeResponse struct {
	// 操作结果。
	Resp           *string `json:"resp,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o AllowDbUserPrivilegeResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "AllowDbUserPrivilegeResponse struct{}"
	}

	return strings.Join([]string{"AllowDbUserPrivilegeResponse", string(data)}, " ")
}

/*
 * DCS
 *
 * DCS V2版本API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type CopyInstanceResponse struct {
	// 备份记录ID。
	BackupId       *string `json:"backup_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o CopyInstanceResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CopyInstanceResponse struct{}"
	}

	return strings.Join([]string{"CopyInstanceResponse", string(data)}, " ")
}

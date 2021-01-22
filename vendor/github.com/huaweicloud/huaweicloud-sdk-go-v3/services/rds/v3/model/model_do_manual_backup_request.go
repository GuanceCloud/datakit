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

// Request Object
type DoManualBackupRequest struct {
	XLanguage *string                    `json:"X-Language,omitempty"`
	Body      *DoManualBackupRequestBody `json:"body,omitempty"`
}

func (o DoManualBackupRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DoManualBackupRequest struct{}"
	}

	return strings.Join([]string{"DoManualBackupRequest", string(data)}, " ")
}

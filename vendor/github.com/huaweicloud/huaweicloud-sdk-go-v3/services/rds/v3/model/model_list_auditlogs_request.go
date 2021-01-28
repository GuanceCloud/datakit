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
type ListAuditlogsRequest struct {
	XLanguage  *string `json:"X-Language,omitempty"`
	InstanceId string  `json:"instance_id"`
	StartTime  string  `json:"start_time"`
	EndTime    string  `json:"end_time"`
	Offset     int32   `json:"offset"`
	Limit      int32   `json:"limit"`
}

func (o ListAuditlogsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListAuditlogsRequest struct{}"
	}

	return strings.Join([]string{"ListAuditlogsRequest", string(data)}, " ")
}

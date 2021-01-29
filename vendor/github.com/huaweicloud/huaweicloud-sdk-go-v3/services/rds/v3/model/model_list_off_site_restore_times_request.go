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
type ListOffSiteRestoreTimesRequest struct {
	XLanguage  *string `json:"X-Language,omitempty"`
	InstanceId string  `json:"instance_id"`
	Date       *string `json:"date,omitempty"`
}

func (o ListOffSiteRestoreTimesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListOffSiteRestoreTimesRequest struct{}"
	}

	return strings.Join([]string{"ListOffSiteRestoreTimesRequest", string(data)}, " ")
}

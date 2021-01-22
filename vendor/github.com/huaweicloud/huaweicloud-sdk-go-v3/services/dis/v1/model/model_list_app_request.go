/*
 * DIS
 *
 * DIS v1 API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ListAppRequest struct {
	Limit        *int32  `json:"limit,omitempty"`
	StartAppName *string `json:"start_app_name,omitempty"`
	StreamName   *string `json:"stream_name,omitempty"`
}

func (o ListAppRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListAppRequest struct{}"
	}

	return strings.Join([]string{"ListAppRequest", string(data)}, " ")
}

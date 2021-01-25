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
type ListAppV3Request struct {
	Limit        *int32  `json:"limit,omitempty"`
	StartAppName *string `json:"start_app_name,omitempty"`
	StreamName   *string `json:"stream_name,omitempty"`
}

func (o ListAppV3Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListAppV3Request struct{}"
	}

	return strings.Join([]string{"ListAppV3Request", string(data)}, " ")
}

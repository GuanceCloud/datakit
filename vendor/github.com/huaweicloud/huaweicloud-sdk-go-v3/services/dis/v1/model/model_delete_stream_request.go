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
type DeleteStreamRequest struct {
	StreamName string `json:"stream_name"`
}

func (o DeleteStreamRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteStreamRequest struct{}"
	}

	return strings.Join([]string{"DeleteStreamRequest", string(data)}, " ")
}

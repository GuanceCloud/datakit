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
type PutRecordsRequest struct {
	Body *PutRecordsRequest `json:"body,omitempty"`
}

func (o PutRecordsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "PutRecordsRequest struct{}"
	}

	return strings.Join([]string{"PutRecordsRequest", string(data)}, " ")
}

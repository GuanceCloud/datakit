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
type ListStreamsRequest struct {
	Limit           *int32  `json:"limit,omitempty"`
	StartStreamName *string `json:"start_stream_name,omitempty"`
}

func (o ListStreamsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListStreamsRequest struct{}"
	}

	return strings.Join([]string{"ListStreamsRequest", string(data)}, " ")
}

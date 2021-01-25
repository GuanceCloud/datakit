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
type UpdatePartitionCountRequest struct {
	StreamName string                       `json:"stream_name"`
	Body       *UpdatePartitionCountRequest `json:"body,omitempty"`
}

func (o UpdatePartitionCountRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdatePartitionCountRequest struct{}"
	}

	return strings.Join([]string{"UpdatePartitionCountRequest", string(data)}, " ")
}

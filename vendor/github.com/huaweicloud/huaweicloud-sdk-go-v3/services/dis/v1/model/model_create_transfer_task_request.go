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
type CreateTransferTaskRequest struct {
	StreamName string                     `json:"stream_name"`
	Body       *CreateTransferTaskRequest `json:"body,omitempty"`
}

func (o CreateTransferTaskRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateTransferTaskRequest struct{}"
	}

	return strings.Join([]string{"CreateTransferTaskRequest", string(data)}, " ")
}

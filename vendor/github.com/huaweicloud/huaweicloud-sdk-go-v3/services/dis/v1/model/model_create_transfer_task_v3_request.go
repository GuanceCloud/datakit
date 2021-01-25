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
type CreateTransferTaskV3Request struct {
	StreamName string                     `json:"stream_name"`
	Body       *CreateTransferTaskRequest `json:"body,omitempty"`
}

func (o CreateTransferTaskV3Request) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateTransferTaskV3Request struct{}"
	}

	return strings.Join([]string{"CreateTransferTaskV3Request", string(data)}, " ")
}

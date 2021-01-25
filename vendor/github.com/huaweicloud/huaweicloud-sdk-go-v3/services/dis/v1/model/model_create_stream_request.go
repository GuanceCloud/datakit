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
type CreateStreamRequest struct {
	Body *CreateStreamReq `json:"body,omitempty"`
}

func (o CreateStreamRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateStreamRequest struct{}"
	}

	return strings.Join([]string{"CreateStreamRequest", string(data)}, " ")
}

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
type CommitCheckpointRequest struct {
	Body *CommitCheckpointRequest `json:"body,omitempty"`
}

func (o CommitCheckpointRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CommitCheckpointRequest struct{}"
	}

	return strings.Join([]string{"CommitCheckpointRequest", string(data)}, " ")
}

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

// Response Object
type CommitCheckpointResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o CommitCheckpointResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CommitCheckpointResponse struct{}"
	}

	return strings.Join([]string{"CommitCheckpointResponse", string(data)}, " ")
}

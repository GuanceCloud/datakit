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
type DeleteCheckpointResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteCheckpointResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteCheckpointResponse struct{}"
	}

	return strings.Join([]string{"DeleteCheckpointResponse", string(data)}, " ")
}

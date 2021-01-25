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
type UpdatePartitionCountResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o UpdatePartitionCountResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdatePartitionCountResponse struct{}"
	}

	return strings.Join([]string{"UpdatePartitionCountResponse", string(data)}, " ")
}

/*
 * SWR
 *
 * SWR API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type UpdateRetentionResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o UpdateRetentionResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateRetentionResponse struct{}"
	}

	return strings.Join([]string{"UpdateRetentionResponse", string(data)}, " ")
}

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
type DeleteRetentionResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteRetentionResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteRetentionResponse struct{}"
	}

	return strings.Join([]string{"DeleteRetentionResponse", string(data)}, " ")
}

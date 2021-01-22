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
type DeleteTriggerResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteTriggerResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteTriggerResponse struct{}"
	}

	return strings.Join([]string{"DeleteTriggerResponse", string(data)}, " ")
}

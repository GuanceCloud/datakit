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
type UpdateTriggerResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o UpdateTriggerResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateTriggerResponse struct{}"
	}

	return strings.Join([]string{"UpdateTriggerResponse", string(data)}, " ")
}

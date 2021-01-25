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
type CreateTriggerResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o CreateTriggerResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateTriggerResponse struct{}"
	}

	return strings.Join([]string{"CreateTriggerResponse", string(data)}, " ")
}

/*
 * BMS
 *
 * BMS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Response Object
type ChangeBaremetalServerNameResponse struct {
	Server         *ChangeBaremetalNameResponsesServers `json:"server,omitempty"`
	HttpStatusCode int                                  `json:"-"`
}

func (o ChangeBaremetalServerNameResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ChangeBaremetalServerNameResponse struct{}"
	}

	return strings.Join([]string{"ChangeBaremetalServerNameResponse", string(data)}, " ")
}

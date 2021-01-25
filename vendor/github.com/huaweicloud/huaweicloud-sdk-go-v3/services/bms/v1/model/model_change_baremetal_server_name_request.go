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

// Request Object
type ChangeBaremetalServerNameRequest struct {
	ServerId string                   `json:"server_id"`
	Body     *ChangeBaremetalNameBody `json:"body,omitempty"`
}

func (o ChangeBaremetalServerNameRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ChangeBaremetalServerNameRequest struct{}"
	}

	return strings.Join([]string{"ChangeBaremetalServerNameRequest", string(data)}, " ")
}

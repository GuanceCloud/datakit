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
type ReinstallBaremetalServerOsRequest struct {
	ServerId string           `json:"server_id"`
	Body     *OsReinstallBody `json:"body,omitempty"`
}

func (o ReinstallBaremetalServerOsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ReinstallBaremetalServerOsRequest struct{}"
	}

	return strings.Join([]string{"ReinstallBaremetalServerOsRequest", string(data)}, " ")
}

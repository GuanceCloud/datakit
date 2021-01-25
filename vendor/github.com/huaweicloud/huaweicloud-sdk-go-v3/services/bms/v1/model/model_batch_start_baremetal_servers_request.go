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
type BatchStartBaremetalServersRequest struct {
	Body *OsStartBody `json:"body,omitempty"`
}

func (o BatchStartBaremetalServersRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchStartBaremetalServersRequest struct{}"
	}

	return strings.Join([]string{"BatchStartBaremetalServersRequest", string(data)}, " ")
}

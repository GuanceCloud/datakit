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
type BatchStopBaremetalServersRequest struct {
	Body *OsStopBody `json:"body,omitempty"`
}

func (o BatchStopBaremetalServersRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "BatchStopBaremetalServersRequest struct{}"
	}

	return strings.Join([]string{"BatchStopBaremetalServersRequest", string(data)}, " ")
}

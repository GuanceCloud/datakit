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
type ShowBaremetalServerVolumeInfoRequest struct {
	ServerId string `json:"server_id"`
}

func (o ShowBaremetalServerVolumeInfoRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowBaremetalServerVolumeInfoRequest struct{}"
	}

	return strings.Join([]string{"ShowBaremetalServerVolumeInfoRequest", string(data)}, " ")
}

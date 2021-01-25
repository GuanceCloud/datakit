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
type ListBareMetalServerDetailsRequest struct {
	ServerId string `json:"server_id"`
}

func (o ListBareMetalServerDetailsRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListBareMetalServerDetailsRequest struct{}"
	}

	return strings.Join([]string{"ListBareMetalServerDetailsRequest", string(data)}, " ")
}

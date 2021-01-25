/*
 * VPC
 *
 * VPC Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

// Request Object
type ShowNetworkIpAvailabilitiesRequest struct {
	NetworkId string `json:"network_id"`
}

func (o ShowNetworkIpAvailabilitiesRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowNetworkIpAvailabilitiesRequest struct{}"
	}

	return strings.Join([]string{"ShowNetworkIpAvailabilitiesRequest", string(data)}, " ")
}

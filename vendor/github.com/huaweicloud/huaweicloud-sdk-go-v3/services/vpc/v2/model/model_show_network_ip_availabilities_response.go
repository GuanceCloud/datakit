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

// Response Object
type ShowNetworkIpAvailabilitiesResponse struct {
	NetworkIpAvailability *NetworkIpAvailability `json:"network_ip_availability,omitempty"`
	HttpStatusCode        int                    `json:"-"`
}

func (o ShowNetworkIpAvailabilitiesResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowNetworkIpAvailabilitiesResponse struct{}"
	}

	return strings.Join([]string{"ShowNetworkIpAvailabilitiesResponse", string(data)}, " ")
}

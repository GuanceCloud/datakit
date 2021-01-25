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
type ShowSubNetworkInterfaceRequest struct {
	SubNetworkInterfaceId string `json:"sub_network_interface_id"`
}

func (o ShowSubNetworkInterfaceRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowSubNetworkInterfaceRequest struct{}"
	}

	return strings.Join([]string{"ShowSubNetworkInterfaceRequest", string(data)}, " ")
}

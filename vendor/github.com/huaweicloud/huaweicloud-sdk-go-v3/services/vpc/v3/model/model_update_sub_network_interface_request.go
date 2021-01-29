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
type UpdateSubNetworkInterfaceRequest struct {
	SubNetworkInterfaceId string                                `json:"sub_network_interface_id"`
	Body                  *UpdateSubNetworkInterfaceRequestBody `json:"body,omitempty"`
}

func (o UpdateSubNetworkInterfaceRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateSubNetworkInterfaceRequest struct{}"
	}

	return strings.Join([]string{"UpdateSubNetworkInterfaceRequest", string(data)}, " ")
}

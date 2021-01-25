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
type UpdateSubNetworkInterfaceResponse struct {
	// 请求ID
	RequestId           *string              `json:"request_id,omitempty"`
	SubNetworkInterface *SubNetworkInterface `json:"sub_network_interface,omitempty"`
	HttpStatusCode      int                  `json:"-"`
}

func (o UpdateSubNetworkInterfaceResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateSubNetworkInterfaceResponse struct{}"
	}

	return strings.Join([]string{"UpdateSubNetworkInterfaceResponse", string(data)}, " ")
}

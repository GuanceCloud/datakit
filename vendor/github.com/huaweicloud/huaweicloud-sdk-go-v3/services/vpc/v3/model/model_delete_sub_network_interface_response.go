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
type DeleteSubNetworkInterfaceResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteSubNetworkInterfaceResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteSubNetworkInterfaceResponse struct{}"
	}

	return strings.Join([]string{"DeleteSubNetworkInterfaceResponse", string(data)}, " ")
}

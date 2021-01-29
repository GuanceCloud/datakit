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
type ShowSubNetworkInterfacesQuantityRequest struct {
}

func (o ShowSubNetworkInterfacesQuantityRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowSubNetworkInterfacesQuantityRequest struct{}"
	}

	return strings.Join([]string{"ShowSubNetworkInterfacesQuantityRequest", string(data)}, " ")
}

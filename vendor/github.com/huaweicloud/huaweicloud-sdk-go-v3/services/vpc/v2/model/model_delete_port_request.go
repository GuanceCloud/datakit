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
type DeletePortRequest struct {
	PortId string `json:"port_id"`
}

func (o DeletePortRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeletePortRequest struct{}"
	}

	return strings.Join([]string{"DeletePortRequest", string(data)}, " ")
}

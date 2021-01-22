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
type ListPortsResponse struct {
	// port列表对象
	Ports          *[]Port `json:"ports,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ListPortsResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ListPortsResponse struct{}"
	}

	return strings.Join([]string{"ListPortsResponse", string(data)}, " ")
}

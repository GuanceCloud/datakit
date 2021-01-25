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
type ShowSubnetResponse struct {
	Subnet         *Subnet `json:"subnet,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o ShowSubnetResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowSubnetResponse struct{}"
	}

	return strings.Join([]string{"ShowSubnetResponse", string(data)}, " ")
}

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
type UpdateVpcPeeringResponse struct {
	Peering        *VpcPeering `json:"peering,omitempty"`
	HttpStatusCode int         `json:"-"`
}

func (o UpdateVpcPeeringResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateVpcPeeringResponse struct{}"
	}

	return strings.Join([]string{"UpdateVpcPeeringResponse", string(data)}, " ")
}

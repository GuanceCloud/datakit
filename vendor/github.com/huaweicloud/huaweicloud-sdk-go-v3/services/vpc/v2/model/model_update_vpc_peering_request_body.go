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

//
type UpdateVpcPeeringRequestBody struct {
	Peering *UpdateVpcPeeringOption `json:"peering"`
}

func (o UpdateVpcPeeringRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateVpcPeeringRequestBody struct{}"
	}

	return strings.Join([]string{"UpdateVpcPeeringRequestBody", string(data)}, " ")
}

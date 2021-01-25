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
type CreateVpcPeeringRequestBody struct {
	Peering *CreateVpcPeeringOption `json:"peering"`
}

func (o CreateVpcPeeringRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateVpcPeeringRequestBody struct{}"
	}

	return strings.Join([]string{"CreateVpcPeeringRequestBody", string(data)}, " ")
}

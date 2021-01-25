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
type RejectVpcPeeringRequest struct {
	PeeringId string `json:"peering_id"`
}

func (o RejectVpcPeeringRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "RejectVpcPeeringRequest struct{}"
	}

	return strings.Join([]string{"RejectVpcPeeringRequest", string(data)}, " ")
}

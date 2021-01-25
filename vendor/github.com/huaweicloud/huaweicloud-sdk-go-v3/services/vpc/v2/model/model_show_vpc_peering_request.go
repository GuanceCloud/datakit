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
type ShowVpcPeeringRequest struct {
	PeeringId string `json:"peering_id"`
}

func (o ShowVpcPeeringRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowVpcPeeringRequest struct{}"
	}

	return strings.Join([]string{"ShowVpcPeeringRequest", string(data)}, " ")
}

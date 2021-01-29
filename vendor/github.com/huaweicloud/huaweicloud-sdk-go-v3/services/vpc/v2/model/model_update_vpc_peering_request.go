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
type UpdateVpcPeeringRequest struct {
	PeeringId string                       `json:"peering_id"`
	Body      *UpdateVpcPeeringRequestBody `json:"body,omitempty"`
}

func (o UpdateVpcPeeringRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateVpcPeeringRequest struct{}"
	}

	return strings.Join([]string{"UpdateVpcPeeringRequest", string(data)}, " ")
}

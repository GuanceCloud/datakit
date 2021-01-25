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
type CreateVpcPeeringRequest struct {
	Body *CreateVpcPeeringRequestBody `json:"body,omitempty"`
}

func (o CreateVpcPeeringRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateVpcPeeringRequest struct{}"
	}

	return strings.Join([]string{"CreateVpcPeeringRequest", string(data)}, " ")
}

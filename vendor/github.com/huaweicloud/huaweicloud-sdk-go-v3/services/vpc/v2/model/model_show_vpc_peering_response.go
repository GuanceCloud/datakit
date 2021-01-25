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
type ShowVpcPeeringResponse struct {
	Peering        *VpcPeering `json:"peering,omitempty"`
	HttpStatusCode int         `json:"-"`
}

func (o ShowVpcPeeringResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowVpcPeeringResponse struct{}"
	}

	return strings.Join([]string{"ShowVpcPeeringResponse", string(data)}, " ")
}

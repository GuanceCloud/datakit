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
type CreateVpcRouteResponse struct {
	Route          *VpcRoute `json:"route,omitempty"`
	HttpStatusCode int       `json:"-"`
}

func (o CreateVpcRouteResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateVpcRouteResponse struct{}"
	}

	return strings.Join([]string{"CreateVpcRouteResponse", string(data)}, " ")
}

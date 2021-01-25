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
type ShowVpcRouteResponse struct {
	Route          *VpcRoute `json:"route,omitempty"`
	HttpStatusCode int       `json:"-"`
}

func (o ShowVpcRouteResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowVpcRouteResponse struct{}"
	}

	return strings.Join([]string{"ShowVpcRouteResponse", string(data)}, " ")
}

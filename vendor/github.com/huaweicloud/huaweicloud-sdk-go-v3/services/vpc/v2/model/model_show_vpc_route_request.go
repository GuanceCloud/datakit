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
type ShowVpcRouteRequest struct {
	RouteId string `json:"route_id"`
}

func (o ShowVpcRouteRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "ShowVpcRouteRequest struct{}"
	}

	return strings.Join([]string{"ShowVpcRouteRequest", string(data)}, " ")
}

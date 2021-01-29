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
type CreateVpcRouteRequest struct {
	Body *CreateVpcRouteRequestBody `json:"body,omitempty"`
}

func (o CreateVpcRouteRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateVpcRouteRequest struct{}"
	}

	return strings.Join([]string{"CreateVpcRouteRequest", string(data)}, " ")
}

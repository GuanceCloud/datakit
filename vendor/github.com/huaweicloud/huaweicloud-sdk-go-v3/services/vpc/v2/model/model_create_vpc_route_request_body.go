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
type CreateVpcRouteRequestBody struct {
	Route *CreateVpcRouteOption `json:"route"`
}

func (o CreateVpcRouteRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateVpcRouteRequestBody struct{}"
	}

	return strings.Join([]string{"CreateVpcRouteRequestBody", string(data)}, " ")
}

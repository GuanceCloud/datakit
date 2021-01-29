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
type CreateVpcRequestBody struct {
	Vpc *CreateVpcOption `json:"vpc"`
}

func (o CreateVpcRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateVpcRequestBody struct{}"
	}

	return strings.Join([]string{"CreateVpcRequestBody", string(data)}, " ")
}

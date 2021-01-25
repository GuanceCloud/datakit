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
type UpdateVpcRequestBody struct {
	Vpc *UpdateVpcOption `json:"vpc"`
}

func (o UpdateVpcRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateVpcRequestBody struct{}"
	}

	return strings.Join([]string{"UpdateVpcRequestBody", string(data)}, " ")
}

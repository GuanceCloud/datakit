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
type UpdateVpcResponse struct {
	Vpc            *Vpc `json:"vpc,omitempty"`
	HttpStatusCode int  `json:"-"`
}

func (o UpdateVpcResponse) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateVpcResponse struct{}"
	}

	return strings.Join([]string{"UpdateVpcResponse", string(data)}, " ")
}

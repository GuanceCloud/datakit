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
type UpdateSubnetRequestBody struct {
	Subnet *UpdateSubnetOption `json:"subnet"`
}

func (o UpdateSubnetRequestBody) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "UpdateSubnetRequestBody struct{}"
	}

	return strings.Join([]string{"UpdateSubnetRequestBody", string(data)}, " ")
}

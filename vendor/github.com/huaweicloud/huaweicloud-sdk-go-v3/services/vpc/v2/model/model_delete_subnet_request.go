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
type DeleteSubnetRequest struct {
	VpcId    string `json:"vpc_id"`
	SubnetId string `json:"subnet_id"`
}

func (o DeleteSubnetRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "DeleteSubnetRequest struct{}"
	}

	return strings.Join([]string{"DeleteSubnetRequest", string(data)}, " ")
}

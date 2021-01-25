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
type CreateVpcRequest struct {
	Body *CreateVpcRequestBody `json:"body,omitempty"`
}

func (o CreateVpcRequest) String() string {
	data, err := json.Marshal(o)
	if err != nil {
		return "CreateVpcRequest struct{}"
	}

	return strings.Join([]string{"CreateVpcRequest", string(data)}, " ")
}
